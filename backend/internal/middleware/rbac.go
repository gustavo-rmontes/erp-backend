package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RBACMiddleware recebe uma lista de roles permitidas e retorna um middleware Gin.
func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtém as claims definidas pelo middleware de autenticação.
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "claims não encontrados"})
			return
		}

		// Faz a conversão das claims para o tipo jwt.MapClaims
		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de claims inválido"})
			return
		}

		// Presume que a role do usuário esteja armazenada na claim "role"
		userRole, exists := mapClaims["role"].(string)
		if !exists || userRole == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "função do usuário não definida"})
			return
		}

		// Verifica se a role do usuário está entre as roles permitidas
		authorized := false
		for _, role := range allowedRoles {
			if strings.EqualFold(userRole, role) {
				authorized = true
				break
			}
		}

		if !authorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acesso negado: permissões insuficientes"})
			return
		}

		c.Next()
	}
}
