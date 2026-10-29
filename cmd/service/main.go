package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Impisigmatus/serial-driver/internal/serial"
	logger "github.com/Impisigmatus/service_core/log"
	"go.bug.st/serial/enumerator"
)

func main() {
	const (
		baud     = 9600
		logLevel = logger.LevelInfo
	)
	log := logger.New(logLevel)

	device, err := getUSBDevice()
	if err != nil {
		log.Panic().Err(err).Msg("Не удалось определить USB устройство")
	}

	port := serial.NewSerialPort(device, baud)
	if err := port.Open(); err != nil {
		log.Panic().Err(err).Msgf("Не удалось открыть %s: %v", device, err)
	}

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
