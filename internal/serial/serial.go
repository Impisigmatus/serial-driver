package serial

import (
	"bufio"
	"sync"

	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
)

type SerialPort struct {
	port   serial.Port
	device string
	baud   int

	lastMessage string
	mu          *sync.RWMutex
}

func NewSerialPort(device string, baud int) *SerialPort {
	return &SerialPort{
		device: device,
		baud:   baud,
		mu:     &sync.RWMutex{},
	}
}

func (sp *SerialPort) Open() error {
	var err error
	sp.port, err = serial.Open(sp.device, &serial.Mode{BaudRate: sp.baud})
	if err != nil {
		return err
	}

	return nil
}

func (sp *SerialPort) Close() error {
	return sp.port.Close()
}

func (sp *SerialPort) Read() error {
	scanner := bufio.NewScanner(sp.port)
	for scanner.Scan() {
		message := scanner.Text()
		sp.mu.Lock()
		sp.lastMessage = message
		sp.mu.Unlock()
		log.Info().Msg(message)
	}
	return scanner.Err()
}

func (sp *SerialPort) GetLastMessage() string {
	sp.mu.RLock()
	message := sp.lastMessage
	sp.mu.RUnlock()
	return message
}
