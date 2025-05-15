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

// GetLogger retorna a instância do logger, inicializando-a se necessário
func GetLogger() *zap.Logger {
	if Logger == nil {
		var err error
		Logger, err = InitLogger()
		if err != nil {
			// Fallback para um logger básico em caso de erro
			Logger, _ = zap.NewProduction()
		}
	}
	return Logger
}

// SetLogger permite definir um logger personalizado (útil para testes)
func SetLogger(l *zap.Logger) {
	Logger = l
}

// WithModule retorna um logger com o campo "module" preenchido
func WithModule(moduleName string) *zap.Logger {
	return GetLogger().With(zap.String("module", moduleName))
}

// SugaredLogger retorna uma versão "sugared" (mais simples de usar) do logger
func SugaredLogger() *zap.SugaredLogger {
	return GetLogger().Sugar()
}

// WithModuleSugared retorna um sugared logger com o campo "module" preenchido
func WithModuleSugared(moduleName string) *zap.SugaredLogger {
	return WithModule(moduleName).Sugar()
}

// Funções auxiliares para não precisar importar zap diretamente

// Field cria um campo de log zap
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// String cria um campo de log zap do tipo string
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Int cria um campo de log zap do tipo int
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Error cria um campo de log zap do tipo error
func Error(key string, value error) zap.Field {
	return zap.Error(value)
}

// Sugar retorna uma versão mais simplificada do logger
func Sugar() *zap.SugaredLogger {
	return Logger.Sugar()
}

// WithError adiciona um campo de erro ao logger
func WithError(err error) *zap.Logger {
	return Logger.With(zap.Error(err))
}
