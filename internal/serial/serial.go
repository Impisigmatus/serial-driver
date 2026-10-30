package serial

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.bug.st/serial"
)

type SerialPort struct {
	logger    zerolog.Logger
	port      serial.Port
	device    string
	baud      int
	timeout   time.Duration
	mu        sync.Mutex
	requestMu sync.Mutex
	pending   chan serialResponse
}

type serialResponse struct {
	message string
	err     error
}

func NewSerialPort(logger zerolog.Logger, baud int) *SerialPort {
	return &SerialPort{
		logger:  logger,
		baud:    baud,
		timeout: 3 * time.Second,
	}
}

func (sp *SerialPort) SetDevice(device string) {
	sp.device = device
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
		if sp.pending != nil {
			sp.pending <- serialResponse{message: message}
			sp.pending = nil
		}
		sp.mu.Unlock()
		sp.logger.Info().Msg(message)
	}
	sp.mu.Lock()
	if sp.pending != nil {
		err := scanner.Err()
		if err == nil {
			err = io.EOF
		}
		sp.pending <- serialResponse{err: err}
		sp.pending = nil
	}
	sp.mu.Unlock()
	return scanner.Err()
}

// SendAndRead sends a newline-terminated message and waits for the next line.
// Read must be running in another goroutine. Only one request may be active.
// Unsolicited lines and late responses cannot be distinguished from replies.
func (sp *SerialPort) SendAndRead(message string) (string, error) {
	if !sp.requestMu.TryLock() {
		return "", errors.New("serial request already in progress")
	}
	defer sp.requestMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), sp.timeout)
	defer cancel()
	response := make(chan serialResponse, 1)
	sp.mu.Lock()
	if sp.port == nil {
		sp.mu.Unlock()
		return "", errors.New("serial port is not open")
	}
	sp.pending = response
	payload := []byte(strings.TrimRight(message, "\r\n") + "\n")
	n, err := sp.port.Write(payload)
	if err == nil && n != len(payload) {
		err = io.ErrShortWrite
	}
	if err != nil {
		sp.pending = nil
	}
	sp.mu.Unlock()
	if err != nil {
		return "", err
	}
	defer func() {
		sp.mu.Lock()
		if sp.pending == response {
			sp.pending = nil
		}
		sp.mu.Unlock()
	}()
	select {
	case reply := <-response:
		return reply.message, reply.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
