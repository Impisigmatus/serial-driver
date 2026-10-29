package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Impisigmatus/serial-driver/autogen/server"
	"github.com/Impisigmatus/serial-driver/internal/config"
	"github.com/Impisigmatus/serial-driver/internal/serial"
	"github.com/Impisigmatus/serial-driver/internal/transport"
	"github.com/Impisigmatus/service_core/log"
	"github.com/Impisigmatus/service_core/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	seriallib "go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// @title Serial Driver API
// @version 1.0.0
// @description Последнее сообщение последовательного порта.
func main() {
	logger := log.New(log.LevelPanic)
	cfg, err := config.Load()
	if err != nil {
		logger.Panic().Err(err).Msg("Invalid config")
	}

	logger = log.New(cfg.LogLevel)
	port := serial.NewSerialPort("", cfg.BaudRate)
	httpServer := &http.Server{Addr: cfg.Port, Handler: getRouter(logger, cfg.SpecPath, port)}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("Не удалось запустить HTTP-сервер")
		}
	}()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer signal.Stop(channel)

	logger.Info().Msg("Service started")
	runSerial(channel, logger, port, cfg.BaudRate)
	if err := httpServer.Shutdown(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Не удалось остановить HTTP-сервер")
	}
	logger.Info().Msg("Service stopped")
}

func getRouter(logger zerolog.Logger, specPath string, port *serial.SerialPort) *chi.Mux {
	srv := transport.New(port)

	router := chi.NewRouter()
	router.Get("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		http.ServeFile(w, r, specPath)
	})
	router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/openapi.yaml")))

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Handle("/api/*",
		middlewares.Use(
			middlewares.Use(
				server.Handler(srv),
				middlewares.ContextLogger()),
			middlewares.RequestID(logger),
		),
	)

	return router
}

func runSerial(channel <-chan os.Signal, log zerolog.Logger, port *serial.SerialPort, baud int) {
	for {
		device, err := getUSBDevice()
		if err != nil {
			log.Warn().Err(err).Msg("Повторный поиск последовательного порта через 5 секунд")
			if !waitRetry(channel) {
				break
			}
			continue
		}

		port.SetDevice(device)
		if err := port.Open(); err != nil {
			log.Warn().Err(err).Msgf("Не удалось открыть %s; повтор через 5 секунд", device)
			if !waitRetry(channel) {
				break
			}
			continue
		}

		readDone := make(chan error, 1)
		go func() {
			log.Info().Msgf("Чтение %s на скорости %d бод", device, baud)
			readDone <- port.Read()
		}()

		select {
		case <-channel:
			if err := port.Close(); err != nil {
				log.Panic().Err(err).Msg("Не удалось закрыть последовательный порт")
			}
			if err := <-readDone; err != nil {
				var portErr *seriallib.PortError
				if !errors.As(err, &portErr) || portErr.Code() != seriallib.PortClosed {
					log.Panic().Err(err).Msg("Ошибка чтения последовательного порта")
				}
			}
			return
		case err := <-readDone:
			log.Warn().Err(err).Msg("Чтение порта завершилось; повторное подключение через 5 секунд")
			if closeErr := port.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Не удалось закрыть последовательный порт")
			}
			if !waitRetry(channel) {
				return
			}
		}
	}
}

func waitRetry(channel <-chan os.Signal) bool {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-channel:
		return false
	}
}

func getUSBDevice() (string, error) {
	devices, err := detectUSBDevices()
	if err != nil {
		return "", fmt.Errorf("Не удалось обнаружить USB устройства: %w", err)
	}

	count := len(devices)
	if count == 0 {
		return "", fmt.Errorf("USB устройство не найдено")
	}

	if count > 1 {
		return "", fmt.Errorf("Найдено несколько USB устройств; невозможно выбрать порт однозначно")
	}

	return devices[0], nil
}

func detectUSBDevices() ([]string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	var devices []string
	for _, candidate := range ports {
		if candidate.IsUSB {
			devices = append(devices, candidate.Name)
		}
	}

	return devices, nil
}
