// Copyright © 2020 The Things Industries B.V.

package logging

import "go.uber.org/zap"

// GetLogger returns a new logger.
func GetLogger(debug bool) *zap.Logger {
	var logger *zap.Logger
	if debug {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Encoding:         "console",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}.Build()
	}
	return logger
}