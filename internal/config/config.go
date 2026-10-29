package config

import (
	"fmt"
	"strings"

	"github.com/Impisigmatus/service_core/log"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
)

type Config struct {
	BaudRate    int       `env:"BAUD_RATE" envDefault:"9600"`
	Port        string    `env:"PORT" envDefault:":8000"`
	SpecPath    string    `env:"OPENAPI_SPEC" envDefault:"autogen/oapi3.yaml"`
	LogLevelStr string    `env:"LOG_LEVEL" envDefault:"debug"`
	LogLevel    log.Level `env:"-"`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	if cfg.BaudRate <= 0 {
		return Config{}, fmt.Errorf("BAUD_RATE должен быть положительным числом")
	}

	level, err := parseLogLevel(cfg.LogLevelStr)
	if err != nil {
		return Config{}, err
	}

	cfg.LogLevel = level
	return cfg, nil
}

func parseLogLevel(value string) (log.Level, error) {
	if strings.EqualFold(value, "warning") {
		value = "warn"
	}
	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return 0, fmt.Errorf("некорректный LOG_LEVEL %q: %w", value, err)
	}
	return log.Level(level), nil
}
