package logger

import "go.uber.org/zap"

func New() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			panic(err)
		}
	}()

	return logger.Sugar(), nil
}
