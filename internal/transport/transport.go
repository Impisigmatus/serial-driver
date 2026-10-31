package transport

import (
	"io"
	"net/http"

	"github.com/Impisigmatus/serial-driver/autogen/server"
	"github.com/Impisigmatus/serial-driver/internal/models"
	"github.com/Impisigmatus/serial-driver/internal/serial"
	"github.com/Impisigmatus/service_core/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
)

type Transport struct {
	port *serial.SerialPort
}

func New(port *serial.SerialPort) server.ServerInterface {
	return &Transport{port: port}
}

// PostApiSerialCommand godoc
//
// @Router /api/serial/command [post]
// @Summary Отправить команду микроконтроллеру
// @Description Отправляет строку команды в последовательный порт и возвращает следующую строку ответа микроконтроллера. Таймаут ожидания ответа — 3 секунды.
// @Accept json
// @Produce json
//
// @Tags Serial
//
// @Param request body models.CommandRequest true "Команда для микроконтроллера в поле data"
// @Success 200 {object} models.CommandResponse "Ответ микроконтроллера в поле data"
// @Failure 400 {string} string "Некорректный JSON-запрос"
// @Failure 500 {string} string "Внутренняя ошибка сервера или ошибка обмена с микроконтроллером"
func (t *Transport) PostApiSerialCommand(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		http.Error(w, "Невалидный логгер", http.StatusInternalServerError)
		return
	}

	var req models.CommandRequest
	requestBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Некорректный JSON-запрос", http.StatusBadRequest)
		return
	}

	if err := jsoniter.Unmarshal(requestBody, &req); err != nil {
		http.Error(w, "Некорректный JSON-запрос", http.StatusBadRequest)
		return
	}

	data, err := t.port.SendAndRead(req.Data)
	if err != nil {
		log.Warn().Err(err).Msg("Не удалось выполнить команду микроконтроллера")
		http.Error(w, "Не удалось выполнить команду микроконтроллера", http.StatusInternalServerError)
		return
	}

	response := &models.CommandResponse{
		Data: data,
	}
	responseBody, err := jsoniter.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось сформировать ответ")
		http.Error(w, "Не удалось сформировать ответ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(responseBody); err != nil {
		log.Error().Err(err).Msg("Не удалось записать ответ")
	}
}
