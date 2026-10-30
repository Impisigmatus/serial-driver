// Этот тестовый файл сгенерирован автоматически.
package serial

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	seriallib "go.bug.st/serial"
)

type replyPort struct {
	seriallib.Port
	reader *io.PipeReader
	writer *io.PipeWriter
	writes chan string
}

func (p *replyPort) Read(b []byte) (int, error) { return p.reader.Read(b) }
func (p *replyPort) Write(b []byte) (int, error) {
	p.writes <- string(b)
	return len(b), nil
}

func TestSendAndRead(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	p := &replyPort{reader: r, writer: w, writes: make(chan string, 1)}
	sp := NewSerialPort(zerolog.Nop(), 9600)
	sp.port = p
	done := make(chan error, 1)
	go func() { done <- sp.Read() }()
	go func() {
		if message := <-p.writes; message != "ping\n" {
			t.Errorf("sent %q", message)
		}
		_, _ = w.Write([]byte("pong\r\n"))
		_ = w.Close()
	}()
	got, err := sp.SendAndRead("ping")
	if err != nil || got != "pong" {
		t.Fatalf("reply = %q, error = %v", got, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestSendAndReadTimeout(t *testing.T) {
	sp := NewSerialPort(zerolog.Nop(), 9600)
	sp.port = &replyPort{writes: make(chan string, 1)}
	sp.timeout = 10 * time.Millisecond
	_, err := sp.SendAndRead("ping")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if sp.pending != nil {
		t.Fatal("pending request was not cleared")
	}
}
