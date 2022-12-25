package mock

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeLoggerClient() plugins.LoggerClient {
	logconfig := zap.NewDevelopmentConfig()
	logconfig.Encoding = "console"
	logconfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zlogger, err := logconfig.Build()
	if err != nil {
		panic(err)
	}
	logger := zapr.NewLogger(zlogger)
	return &fakeLoggerClient{logger: logger}
}

type fakeLoggerClient struct {
	logger logr.Logger
}

func (c *fakeLoggerClient) GetLogger() logr.Logger {
	return c.logger
}
