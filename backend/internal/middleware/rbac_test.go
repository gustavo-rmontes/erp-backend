package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestRBACMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Define uma rota protegida com RBAC para a role "admin"
	router.Use(func(c *gin.Context) {
		// Simula que o middleware de autenticação já inseriu as claims no contexto
		claims := jwt.MapClaims{
			"role": "user", // Alterar para "admin" para testar a autorização positiva
		}
		c.Set("claims", claims)
		c.Next()
	})
	router.Use(RBACMiddleware("admin"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "acesso concedido"})
	})

	// Testa acesso com role "user" (não autorizado)
	req, _ := http.NewRequest("GET", "/protected", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Errorf("esperado 403, obtido %d", resp.Code)
	}
}
