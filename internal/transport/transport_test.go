// Этот тестовый файл сгенерирован автоматически.
package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Impisigmatus/serial-driver/internal/serial"
	"github.com/Impisigmatus/service_core/log"
	"github.com/rs/zerolog"
)

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("body read failed") }

func TestPostApiSerialCommand(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       io.Reader
		noLogger   bool
		wantStatus int
		wantBody   string
	}{
		{name: "invalid JSON", body: strings.NewReader(`{`), wantStatus: http.StatusBadRequest, wantBody: "Некорректный JSON-запрос\n"},
		{name: "empty body", body: strings.NewReader(""), wantStatus: http.StatusBadRequest, wantBody: "Некорректный JSON-запрос\n"},
		{name: "invalid data type", body: strings.NewReader(`{"data":42}`), wantStatus: http.StatusBadRequest, wantBody: "Некорректный JSON-запрос\n"},
		{name: "body too large", body: strings.NewReader(strings.Repeat("x", (1<<20)+1)), wantStatus: http.StatusBadRequest, wantBody: "Некорректный JSON-запрос\n"},
		{name: "body read failure", body: failingBody{}, wantStatus: http.StatusBadRequest, wantBody: "Некорректный JSON-запрос\n"},
		{name: "port not open", body: strings.NewReader(`{"data":"RELEASE"}`), wantStatus: http.StatusInternalServerError, wantBody: "Не удалось выполнить команду микроконтроллера\n"},
		{name: "missing logger", body: strings.NewReader(`{"data":"RELEASE"}`), noLogger: true, wantStatus: http.StatusInternalServerError, wantBody: "Невалидный логгер\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := zerolog.New(&logs)
			request := httptest.NewRequest(http.MethodPost, "/api/serial/command", tc.body)
			if !tc.noLogger {
				request = request.WithContext(context.WithValue(request.Context(), log.CtxKey, logger))
			}
			recorder := httptest.NewRecorder()
			New(serial.NewSerialPort(logger, 9600)).PostApiSerialCommand(recorder, request)
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}
			if got := recorder.Body.String(); got != tc.wantBody {
				t.Fatalf("body = %q, want %q", got, tc.wantBody)
			}
			if tc.name == "port not open" && !strings.Contains(logs.String(), "serial port is not open") {
				t.Fatalf("serial error missing from log: %s", logs.String())
			}
		})
	}
}
