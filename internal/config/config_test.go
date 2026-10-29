// Этот тестовый файл сгенерирован автоматически.
package config

import (
	"strings"
	"testing"

	"github.com/Impisigmatus/service_core/log"
)

func TestLoadDefaults(t *testing.T) {
	for _, name := range []string{"BAUD_RATE", "PORT", "OPENAPI_SPEC", "LOG_LEVEL"} {
		t.Setenv(name, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaudRate != 9600 || cfg.Port != ":8000" || cfg.SpecPath != "autogen/oapi3.yaml" || cfg.LogLevel != log.LevelDebug {
		t.Fatalf("неверные значения по умолчанию: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("BAUD_RATE", "115200")
	t.Setenv("PORT", ":9000")
	t.Setenv("OPENAPI_SPEC", "/tmp/spec.yaml")
	t.Setenv("LOG_LEVEL", "warning")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaudRate != 115200 || cfg.Port != ":9000" || cfg.SpecPath != "/tmp/spec.yaml" || cfg.LogLevel != log.LevelWarning {
		t.Fatalf("переменные окружения не применены: %+v", cfg)
	}
}

func TestLoadInvalidBaudRate(t *testing.T) {
	for _, value := range []string{"invalid", "0", "-1"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("BAUD_RATE", value)
			if _, err := Load(); err == nil {
				t.Fatalf("ожидалась ошибка BAUD_RATE, получено: %v", err)
			}
		})
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "invalid")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "LOG_LEVEL") {
		t.Fatalf("ожидалась ошибка LOG_LEVEL, получено: %v", err)
	}
}
