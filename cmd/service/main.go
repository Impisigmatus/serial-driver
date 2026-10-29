package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Impisigmatus/serial-driver/internal/serial"
	logger "github.com/Impisigmatus/service_core/log"
)

func main() {
	const (
		device   = "/dev/ttyACM0"
		baud     = 9600
		logLevel = logger.LevelInfo
	)
	log := logger.New(logLevel)

	port := serial.NewSerialPort(device, baud)
	if err := port.Open(); err != nil {
		log.Panic().Err(err).Msgf("Не удалось открыть %s: %v", device, err)
	}
	defer port.Close()

	go func() {
		log.Info().Msgf("Чтение %s на скорости %d бод", device, baud)
		if err := port.Read(); err != nil {
			log.Panic().Err(err).Msg("Invalid service starting")
		}
		log.Info().Msg("Service stopped")
	}()
	log.Info().Msg("Service started")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-channel

	if err := port.Close(); err != nil {
		log.Panic().Err(err).Msg("Invalid service stopping")
	}
}
