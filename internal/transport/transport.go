package transport

import (
	"net/http"

	"github.com/Impisigmatus/serial-driver/autogen/server"
	"github.com/Impisigmatus/service_core/log"
	"github.com/rs/zerolog"
)

type messageReader interface {
	GetLastMessage() string
}

type Transport struct {
	reader messageReader
}

func New(reader messageReader) server.ServerInterface {
	return &Transport{reader: reader}
}

// GetApiSerialLatest godoc
//
// @Router /api/serial/latest [get]
// @Summary Последнее сообщение последовательного порта
// @Description Возвращает последнюю полную строку, полученную от устройства.
// @Produce plain
//
// @Tags Serial
//
// @Success 200 {string} string "Последнее сообщение"
// @Failure 500 {string} string "Невалидный логгер"
// @Failure 503 {string} string "Сообщение ещё не получено"
func (t *Transport) GetApiSerialLatest(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		http.Error(w, "Невалидный логгер", http.StatusInternalServerError)
		return
	}

	message := t.reader.GetLastMessage()
	if message == "" {
		log.Warn().Msg("Сообщение ещё не получено")
		http.Error(w, "Сообщение ещё не получено", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(message)); err != nil {
		log.Error().Err(err).Msg("Не удалось записать ответ")
	}
}
