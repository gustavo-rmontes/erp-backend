package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestAuditMiddleware(t *testing.T) {
	// Inicializa o logger para teste (modo de desenvolvimento)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Configura o Gin em modo de teste
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Aplica o AuditMiddleware
	router.Use(AuditMiddleware(logger))

	// Define uma rota de teste
	router.GET("/test", func(c *gin.Context) {
		// Simula definir o usuário no contexto (geralmente definido pelo AuthMiddleware)
		c.Set("user", "teste_user")
		time.Sleep(10 * time.Millisecond) // Simula algum processamento
		c.JSON(http.StatusOK, gin.H{"message": "rota de teste"})
	})

	// Cria uma requisição de teste
	req, _ := http.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verifica se a resposta foi 200 OK
	if resp.Code != http.StatusOK {
		t.Errorf("esperado 200, obtido %d", resp.Code)
	}
}
