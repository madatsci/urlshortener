// Package logger provides structured and leveled logging.
package logger

import "go.uber.org/zap"

// New builds a logger.
func New() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	defer logger.Sync() //nolint:errcheck

	return logger.Sugar(), nil
}
