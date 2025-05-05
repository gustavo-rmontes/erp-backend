package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ERP-ONSMART/backend/internal/modules/products/models"

	"github.com/gin-gonic/gin"
)

// getValidProductID cria um produto via endpoint e retorna seu ID para associar à garantia.
func getValidProductID(t *testing.T) int {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/products", CreateProductHandler)
	router.GET("/products", ListProductsHandler)

	// Cria o produto
	productBody := []byte(`{
		"name": "Produto para Warranty",
		"description": "Produto necessário para testar warranty",
		"price": 150.00,
		"stock": 20
	}`)
	reqCreate, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(productBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	// Não esperamos o ID na resposta, então vamos listar os produtos para encontrar o último inserido.
	reqList, _ := http.NewRequest("GET", "/products", nil)
	respList := httptest.NewRecorder()
	router.ServeHTTP(respList, reqList)

	var listResult struct {
		Products []struct {
			ID int `json:"id"`
		} `json:"products"`
	}
	if err := json.Unmarshal(respList.Body.Bytes(), &listResult); err != nil {
		t.Fatalf("Erro ao decodificar resposta de produtos: %v", err)
	}
	if len(listResult.Products) == 0 {
		t.Fatal("Nenhum produto encontrado após criação")
	}
	return listResult.Products[len(listResult.Products)-1].ID
}

func TestCreateWarrantyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Registra endpoints necessários
	router.POST("/products", CreateProductHandler)
	router.GET("/products", ListProductsHandler)
	router.POST("/warranties", CreateWarrantyHandler)

	// Cria um produto para associar a garantia e obtém seu ID
	productID := getValidProductID(t)

	// Prepara o JSON da warranty com o product_id válido
	warrantyPayload := map[string]interface{}{
		"product_id":      productID,
		"duration_months": 12,
		"price":           25.50,
	}
	payloadBytes, err := json.Marshal(warrantyPayload)
	if err != nil {
		t.Fatalf("Erro ao fazer marshal da warranty: %v", err)
	}

	req, _ := http.NewRequest("POST", "/warranties", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado 201, obtido %d", resp.Code)
	}
}

func TestListWarrantiesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/warranties", ListWarrantiesHandler)

	req, _ := http.NewRequest("GET", "/warranties", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", resp.Code)
	}

	// Decodifica resposta para verificar se há warranties
	var result struct {
		Warranties []models.Warranty `json:"warranties"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Erro ao decodificar resposta de warranties: %v", err)
	}
	// Para o teste, espera-se que haja pelo menos uma garantia
	if len(result.Warranties) == 0 {
		t.Errorf("Nenhuma garantia retornada")
	}
}

func TestUpdateWarrantyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Registra os endpoints necessários
	router.POST("/products", CreateProductHandler)
	router.GET("/products", ListProductsHandler)
	router.POST("/warranties", CreateWarrantyHandler)
	router.PUT("/warranties/:id", UpdateWarrantyHandler)
	router.GET("/warranties", ListWarrantiesHandler)

	// Cria um produto para associar à garantia
	productID := getValidProductID(t)

	// Cria uma garantia
	warrantyPayload := map[string]interface{}{
		"product_id":      productID,
		"duration_months": 6,
		"price":           30.0,
	}
	payloadBytes, _ := json.Marshal(warrantyPayload)
	reqCreate, _ := http.NewRequest("POST", "/warranties", bytes.NewBuffer(payloadBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	if respCreate.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar garantia, código: %d", respCreate.Code)
	}

	// Lista warranties e obtém o ID da última inserida
	reqList, _ := http.NewRequest("GET", "/warranties", nil)
	respList := httptest.NewRecorder()
	router.ServeHTTP(respList, reqList)

	var listResult struct {
		Warranties []struct {
			ID int `json:"id"`
		} `json:"warranties"`
	}
	if err := json.Unmarshal(respList.Body.Bytes(), &listResult); err != nil {
		t.Fatalf("Erro ao decodificar warranties: %v", err)
	}
	if len(listResult.Warranties) == 0 {
		t.Fatal("Nenhuma warranty encontrada para atualização")
	}
	warrantyID := listResult.Warranties[len(listResult.Warranties)-1].ID

	// Prepara os dados atualizados para warranty
	updatedPayload := map[string]interface{}{
		"product_id":      productID,
		"duration_months": 12,
		"price":           35.75,
	}
	updatedBytes, _ := json.Marshal(updatedPayload)
	reqUpdate, _ := http.NewRequest("PUT", "/warranties/"+strconv.Itoa(warrantyID), bytes.NewBuffer(updatedBytes))
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate := httptest.NewRecorder()
	router.ServeHTTP(respUpdate, reqUpdate)

	if respUpdate.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", respUpdate.Code)
	}
}

func TestDeleteWarrantyHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Registra endpoints necessários
	router.POST("/products", CreateProductHandler)
	router.GET("/products", ListProductsHandler)
	router.POST("/warranties", CreateWarrantyHandler)
	router.DELETE("/warranties/:id", DeleteWarrantyHandler)
	router.GET("/warranties", ListWarrantiesHandler)

	// Cria um produto para a garantia
	productID := getValidProductID(t)

	// Cria uma garantia
	warrantyPayload := map[string]interface{}{
		"product_id":      productID,
		"duration_months": 6,
		"price":           28.0,
	}
	payloadBytes, _ := json.Marshal(warrantyPayload)
	reqCreate, _ := http.NewRequest("POST", "/warranties", bytes.NewBuffer(payloadBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	if respCreate.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar warranty para deleção, código: %d", respCreate.Code)
	}

	// Lista warranties para obter o ID da última inserida
	reqList, _ := http.NewRequest("GET", "/warranties", nil)
	respList := httptest.NewRecorder()
	router.ServeHTTP(respList, reqList)
	var listResult struct {
		Warranties []struct {
			ID int `json:"id"`
		} `json:"warranties"`
	}
	if err := json.Unmarshal(respList.Body.Bytes(), &listResult); err != nil {
		t.Fatalf("Erro ao decodificar warranties: %v", err)
	}
	if len(listResult.Warranties) == 0 {
		t.Fatal("Nenhuma warranty encontrada para deleção")
	}
	warrantyID := listResult.Warranties[len(listResult.Warranties)-1].ID

	// Deleta a warranty
	reqDelete, _ := http.NewRequest("DELETE", "/warranties/"+strconv.Itoa(warrantyID), nil)
	respDelete := httptest.NewRecorder()
	router.ServeHTTP(respDelete, reqDelete)
	if respDelete.Code != http.StatusOK {
		t.Errorf("Esperado 200 ao deletar warranty, obtido %d", respDelete.Code)
	}
}
