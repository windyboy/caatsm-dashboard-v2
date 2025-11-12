package observability

import (
	"fmt"
	"strings"

	"github.com/windy/caatsm-dashboard/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a zap.Logger based on application configuration.
func NewLogger(cfg config.LoggerConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	if strings.EqualFold(cfg.Format, "console") {
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	if !cfg.EnableStacktrace {
		zapCfg.DisableStacktrace = true
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	return logger, nil
}
