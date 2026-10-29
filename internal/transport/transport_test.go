package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Impisigmatus/service_core/log"
)

type stubReader struct{ message string }

func (r stubReader) GetLastMessage() string { return r.message }

func TestGetApiSerialLatest(t *testing.T) {
	for _, tc := range []struct {
		name, message string
		wantStatus    int
	}{
		{"no message", "", http.StatusServiceUnavailable},
		{"latest message", "Привет из Arduino!", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/serial/latest", nil)
			request = request.WithContext(context.WithValue(request.Context(), log.CtxKey, log.New(log.LevelInfo)))
			New(stubReader{tc.message}).GetApiSerialLatest(recorder, request)
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK {
				if got := recorder.Body.String(); got != tc.message {
					t.Fatalf("message = %q, want %q", got, tc.message)
				}
				if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
					t.Fatalf("Content-Type = %q", got)
				}
			}
		})
	}
}
