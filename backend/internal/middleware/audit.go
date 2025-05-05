package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditMiddleware registra informações de auditoria para cada requisição.
func AuditMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Captura o tempo inicial
		startTime := time.Now()

		// Processa a requisição
		c.Next()

		// Calcula a latência
		duration := time.Since(startTime)

		// Recupera informações do usuário, se definidas (p.ex.: pelo AuthMiddleware)
		user, exists := c.Get("user")
		if !exists {
			user = "desconhecido"
		}

		// Registra a auditoria com detalhes da requisição
		logger.Info("Audit Log",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", duration),
			zap.Any("user", user),
		)
	}
}
