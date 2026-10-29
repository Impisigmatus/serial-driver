package main

import (
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

	log.Info().Msgf("Чтение %s на скорости %d бод", device, baud)
	port.Read()
}
