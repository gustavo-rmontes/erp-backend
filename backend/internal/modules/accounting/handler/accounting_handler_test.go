package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

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

	// Inicializa o logger para evitar panic em testes
	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

func TestCreateTransactionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/accounting", CreateTransactionHandler)

	body := []byte(`{
		"description": "Compra de insumos",
		"amount": 1234.56,
		"date": "08/04/2025"
	}`)

	req, _ := http.NewRequest("POST", "/accounting", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado status 201, obtido %d", resp.Code)
	}
}

func TestListTransactionsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/accounting", ListTransactionsHandler)

	req, _ := http.NewRequest("GET", "/accounting", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", resp.Code)
	}

	// Valida se a resposta possui o campo "data"
	var result struct {
		Data []struct {
			ID          int     `json:"id"`
			Description string  `json:"description"`
			Amount      float64 `json:"amount"`
			Date        string  `json:"date"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Errorf("Erro ao decodificar resposta: %v", err)
	}
}

func TestUpdateTransactionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Cria as rotas necessárias
	router.POST("/accounting", CreateTransactionHandler)
	router.PUT("/accounting/:id", UpdateTransactionHandler)
	router.GET("/accounting", ListTransactionsHandler)

	// Cria uma transação de teste
	createBody := []byte(`{
		"description": "Atualizar transação",
		"amount": 100.50,
		"date": "08/04/2025"
	}`)
	req, _ := http.NewRequest("POST", "/accounting", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, req)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("Falha ao criar transação: %d", createResp.Code)
	}

	// Recupera o ID da transação criada via GET
	listResp := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/accounting", nil)
	router.ServeHTTP(listResp, reqList)
	var result struct {
		Data []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}
	if len(result.Data) == 0 {
		t.Fatalf("Nenhuma transação retornada")
	}
	id := result.Data[len(result.Data)-1].ID

	// Atualiza a transação criada
	updateBody := []byte(`{
		"description": "Transação atualizada",
		"amount": 200.75,
		"date": "09/04/2025"
	}`)
	reqUpdate, _ := http.NewRequest("PUT", "/accounting/"+strconv.Itoa(id), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, reqUpdate)
	if updateResp.Code != http.StatusOK {
		t.Errorf("Esperado status 200 na atualização, obtido %d", updateResp.Code)
	}
}

func TestDeleteTransactionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Cria as rotas necessárias
	router.POST("/accounting", CreateTransactionHandler)
	router.DELETE("/accounting/:id", DeleteTransactionHandler)
	router.GET("/accounting", ListTransactionsHandler)

	// Cria uma transação para ser deletada
	body := []byte(`{
		"description": "Para deletar",
		"amount": 500,
		"date": "08/04/2025"
	}`)
	req, _ := http.NewRequest("POST", "/accounting", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, req)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar transação para deletar: %d", createResp.Code)
	}

	// Recupera o ID da última transação criada
	listResp := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/accounting", nil)
	router.ServeHTTP(listResp, reqList)
	var result struct {
		Data []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}
	if len(result.Data) == 0 {
		t.Fatalf("Nenhuma transação encontrada para deletar")
	}
	id := result.Data[len(result.Data)-1].ID

	// Deleta a transação
	reqDel, _ := http.NewRequest("DELETE", "/accounting/"+strconv.Itoa(id), nil)
	delResp := httptest.NewRecorder()
	router.ServeHTTP(delResp, reqDel)
	if delResp.Code != http.StatusOK {
		t.Errorf("Esperado status 200 ao deletar, obtido %d", delResp.Code)
	}
}
