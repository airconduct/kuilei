package pluginhelpers

import (
	"github.com/go-logr/logr"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func MakeLoggerClient(getLogger func() logr.Logger) plugins.LoggerClient {
	return getLoggerFunc(getLogger)
}

type getLoggerFunc func() logr.Logger

func (fn getLoggerFunc) GetLogger() logr.Logger {
	return fn()
}
