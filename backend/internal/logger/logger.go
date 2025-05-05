package logger

import (
	"go.uber.org/zap"
)

// Logger é a instância global do logger.
var Logger *zap.Logger

// InitLogger inicializa a instância global do logger.
// Aqui, usamos o modo de desenvolvimento para uma saída mais amigável durante o desenvolvimento.
// Para produção, considere usar zap.NewProduction().
func InitLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	Logger = logger
	return logger, nil
}
