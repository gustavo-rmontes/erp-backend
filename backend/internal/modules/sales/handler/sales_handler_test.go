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

func TestCreateSaleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/sales", CreateSaleHandler)

	// Exemplo de JSON com dados válidos
	body := []byte(`{
		"product": "Produto A",
		"quantity": 5,
		"price": 123.45,
		"customer": "cliente@example.com"
	}`)

	req, _ := http.NewRequest("POST", "/sales", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado status 201, obtido %d", resp.Code)
	}
}

func TestListSalesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/sales", ListSalesHandler)

	req, _ := http.NewRequest("GET", "/sales", nil)
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

func TestGetSaleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup routes
	router.POST("/sales", CreateSaleHandler)
	router.GET("/sales/:id", GetSaleHandler)

	// 1. Create a test sale
	createBody := []byte(`{
		"product": "Produto para GetSale Handler",
		"quantity": 7,
		"price": 75.99,
		"customer": "getsale@handler.com"
	}`)

	req, _ := http.NewRequest("POST", "/sales", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("Falha ao criar venda de teste: código %d", resp.Code)
	}

	// 2. Extract the ID of the created sale
	var createdSale struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &createdSale); err != nil {
		t.Fatalf("Erro ao decodificar resposta da criação: %v", err)
	}

	// 3. Test successful retrieval
	getReq, _ := http.NewRequest("GET", "/sales/"+strconv.Itoa(createdSale.ID), nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Errorf("Esperado status 200 ao obter venda, obtido %d", getResp.Code)
	}

	// 4. Verify the returned sale data
	var retrievedSale struct {
		ID       int     `json:"id"`
		Product  string  `json:"product"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
		Customer string  `json:"customer"`
	}
	if err := json.Unmarshal(getResp.Body.Bytes(), &retrievedSale); err != nil {
		t.Fatalf("Erro ao decodificar venda recuperada: %v", err)
	}

	if retrievedSale.ID != createdSale.ID {
		t.Errorf("ID incorreto, esperado %d, obtido %d", createdSale.ID, retrievedSale.ID)
	}

	// 5. Test non-existent sale
	nonExistingID := 99999
	nonExistReq, _ := http.NewRequest("GET", "/sales/"+strconv.Itoa(nonExistingID), nil)
	nonExistResp := httptest.NewRecorder()
	router.ServeHTTP(nonExistResp, nonExistReq)

	if nonExistResp.Code != http.StatusNotFound {
		t.Errorf("Esperado status 404 para venda inexistente, obtido %d", nonExistResp.Code)
	}

	// 6. Test invalid ID format
	invalidReq, _ := http.NewRequest("GET", "/sales/abc", nil)
	invalidResp := httptest.NewRecorder()
	router.ServeHTTP(invalidResp, invalidReq)

	if invalidResp.Code != http.StatusBadRequest {
		t.Errorf("Esperado status 400 para ID inválido, obtido %d", invalidResp.Code)
	}
}

func TestUpdateSaleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Cria as rotas necessárias
	router.POST("/sales", CreateSaleHandler)
	router.PUT("/sales/:id", UpdateSaleHandler)
	router.GET("/sales", ListSalesHandler)

	// Cria uma venda de teste com dados válidos
	createBody := []byte(`{
		"product": "Produto A",
		"quantity": 10,
		"price": 100.50,
		"customer": "cliente@example.com"
	}`)
	req, _ := http.NewRequest("POST", "/sales", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, req)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("Falha ao criar venda: %d", createResp.Code)
	}

	// Recupera o ID da venda criada via GET
	listResp := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/sales", nil)
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
		t.Fatalf("Nenhuma venda retornada")
	}
	id := result.Data[len(result.Data)-1].ID

	// Atualiza a venda criada
	updateBody := []byte(`{
		"product": "Produto Atualizado",
		"quantity": 5,
		"price": 200.75,
		"customer": "cliente@exemplo.com"
	}`)
	reqUpdate, _ := http.NewRequest("PUT", "/sales/"+strconv.Itoa(id), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, reqUpdate)
	if updateResp.Code != http.StatusOK {
		t.Errorf("Esperado status 200 na atualização, obtido %d", updateResp.Code)
	}
}

func TestDeleteSaleHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Cria as rotas necessárias
	router.POST("/sales", CreateSaleHandler)
	router.DELETE("/sales/:id", DeleteSaleHandler)
	router.GET("/sales", ListSalesHandler)

	// Cria uma venda para ser deletada
	body := []byte(`{
		"product": "Produto para deletar",
		"quantity": 5,
		"price": 500,
		"customer": "cliente@exemplo.com"
	}`)
	req, _ := http.NewRequest("POST", "/sales", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, req)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar venda para deletar: %d", createResp.Code)
	}

	// Recupera o ID da última venda criada
	listResp := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/sales", nil)
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
		t.Fatalf("Nenhuma venda encontrada para deletar")
	}
	id := result.Data[len(result.Data)-1].ID

	// Deleta a venda
	reqDel, _ := http.NewRequest("DELETE", "/sales/"+strconv.Itoa(id), nil)
	delResp := httptest.NewRecorder()
	router.ServeHTTP(delResp, reqDel)
	if delResp.Code != http.StatusOK {
		t.Errorf("Esperado status 200 ao deletar, obtido %d", delResp.Code)
	}
}
