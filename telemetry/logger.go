package telemetry

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = *zap.Logger

type LoggerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

func NewLogger(config LoggerConfig) (Logger, error) {
	return zap.NewProduction(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.Fields(
			zap.String("serviceName", config.ServiceName),
			zap.String("serviceVersion", config.ServiceVersion),
			zap.String("environment", config.Environment),
		),
	)
}

func Any(k string, v any) zap.Field {
	return zap.Any(k, v)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func String(k, v string) zap.Field {
	return zap.String(k, v)
}
