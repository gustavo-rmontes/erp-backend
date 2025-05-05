package handler

import (
	"ERP-ONSMART/backend/internal/modules/auth/models"
	"ERP-ONSMART/backend/internal/modules/auth/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func LoginHandler(c *gin.Context) {
	var creds models.LoginRequest
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	user, err := service.Authenticate(creds.Username, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jwtSecret := viper.GetString("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(2 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login realizado com sucesso", "token": tokenStr})
}

func RegisterHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	if err := service.Register(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usuário registrado com sucesso"})
}

func ProfileHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não fornecido"})
		return
	}
	var tokenString string
	_, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString)
	if err != nil || tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token malformado"})
		return
	}
	jwtSecret := viper.GetString("JWT_SECRET")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["username"] == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims inválidas"})
		return
	}
	user, err := service.GetUserProfile(claims["username"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar perfil"})
		return
	}

	// Retorna apenas o nome
	c.JSON(http.StatusOK, gin.H{
		"message": "Perfil do usuário",
		"nome":    user.Nome,
	})
}

func DeleteUserHandler(c *gin.Context) {
	username := c.Param("username")

	if err := service.DeleteUser(username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao deletar usuário",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário deletado com sucesso!"})
}
