package nowire

import (
	"go.uber.org/zap"
)

func NewZapLogger(debug bool) *zap.Logger {
	logger, err := (func() (*zap.Logger, error) {
		if debug {
			return zap.NewDevelopment()
		} else {
			return zap.NewProduction()
		}
	})()
	if err != nil {
		panic(err)
	}

	return logger
}
