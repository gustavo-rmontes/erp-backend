package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ERP-ONSMART/backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}

	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

func TestRegisterAndLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/auth/register", RegisterHandler)
	r.POST("/auth/login", LoginHandler)

	username := "testuser_" + time.Now().Format("150405")
	password := "123456"

	body := []byte(`{"username":"` + username + `","password":"` + password + `"}`)

	// Testa registro
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("Erro ao registrar usuário. Código: %d", resp.Code)
	}

	// Testa login
	reqLogin, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	reqLogin.Header.Set("Content-Type", "application/json")
	respLogin := httptest.NewRecorder()
	r.ServeHTTP(respLogin, reqLogin)
	if respLogin.Code != http.StatusOK {
		t.Errorf("Erro ao logar usuário. Código: %d", respLogin.Code)
	}

	// Verifica presença do token na resposta
	var result map[string]interface{}
	json.Unmarshal(respLogin.Body.Bytes(), &result)
	if _, ok := result["token"]; !ok {
		t.Error("Token não retornado no login")
	}
}

func TestProfileHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/auth/register", RegisterHandler)
	r.POST("/auth/login", LoginHandler)
	r.GET("/auth/profile", ProfileHandler)

	username := "testprofile_" + time.Now().Format("150405")
	password := "senhaSegura"
	body := []byte(`{"username":"` + username + `","password":"` + password + `"}`)

	// Registra usuário
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	// Faz login
	reqLogin, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	reqLogin.Header.Set("Content-Type", "application/json")
	respLogin := httptest.NewRecorder()
	r.ServeHTTP(respLogin, reqLogin)

	var loginResp map[string]interface{}
	json.Unmarshal(respLogin.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	// Chama /profile com o token
	reqProfile, _ := http.NewRequest("GET", "/auth/profile", nil)
	reqProfile.Header.Set("Authorization", "Bearer "+token)
	respProfile := httptest.NewRecorder()
	r.ServeHTTP(respProfile, reqProfile)

	if respProfile.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", respProfile.Code)
	}
}

func TestDeleteUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/auth/register", RegisterHandler)
	r.DELETE("/auth/:username", DeleteUserHandler)

	username := "testdelete_" + time.Now().Format("150405")
	password := "senha123"
	body := []byte(`{"username":"` + username + `","password":"` + password + `"}`)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	reqDel, _ := http.NewRequest("DELETE", "/auth/"+username, nil)
	respDel := httptest.NewRecorder()
	r.ServeHTTP(respDel, reqDel)

	if respDel.Code != http.StatusOK {
		t.Errorf("Esperado 200 ao deletar, obtido %d", respDel.Code)
	}
}
