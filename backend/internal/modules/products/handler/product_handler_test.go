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
	models "ERP-ONSMART/backend/internal/modules/products/models"

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

	// Inicia logger global
	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao iniciar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

func TestCreateProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/products", CreateProductHandler)

	product := models.ProductToAdd
	body, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Erro ao converter ProductToAdd para JSON: %v", err)
	}

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado 201, obtido %d", resp.Code)
	}
}

func TestListProductsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/products", ListProductsHandler)

	req, _ := http.NewRequest("GET", "/products", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", resp.Code)
	}
}

func TestGetProductByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/products", CreateProductHandler)
	r.GET("/products/:id", GetProductByIDHandler)

	// Cria um produto usando ProductToAdd
	product := models.ProductToAdd
	body, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Erro ao converter ProductToAdd para JSON: %v", err)
	}

	reqCreate, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	r.ServeHTTP(respCreate, reqCreate)

	if respCreate.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar produto. Esperado 201, obtido %d", respCreate.Code)
	}

	// Extrai o ID do produto criado
	var createdProduct models.Product
	if err := json.Unmarshal(respCreate.Body.Bytes(), &createdProduct); err != nil {
		t.Fatalf("Erro ao desserializar resposta do produto criado: %v", err)
	}

	// Busca o produto pelo ID usando o handler
	reqGet, _ := http.NewRequest("GET", "/products/"+strconv.Itoa(createdProduct.ID), nil)
	respGet := httptest.NewRecorder()
	r.ServeHTTP(respGet, reqGet)

	if respGet.Code != http.StatusOK {
		t.Fatalf("Erro ao buscar produto por ID. Esperado 200, obtido %d", respGet.Code)
	}

	// Verifica se o produto retornado é o esperado
	var fetchedProduct struct {
		Product models.Product `json:"product"`
	}
	if err := json.Unmarshal(respGet.Body.Bytes(), &fetchedProduct); err != nil {
		t.Fatalf("Erro ao desserializar resposta do produto buscado: %v", err)
	}

	if fetchedProduct.Product.ID != createdProduct.ID {
		t.Errorf("Produto retornado não corresponde ao esperado. Esperado ID: %d, Obtido ID: %d", createdProduct.ID, fetchedProduct.Product.ID)
	}
}

func TestUpdateProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/products", CreateProductHandler)
	r.PUT("/products/:id", UpdateProductHandler)
	r.GET("/products", ListProductsHandler)

	// Cria produto usando ProductToAdd
	product := models.ProductToAdd
	body, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Erro ao converter ProductToAdd para JSON: %v", err)
	}

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Lista produtos
	reqList, _ := http.NewRequest("GET", "/products", nil)
	respList := httptest.NewRecorder()
	r.ServeHTTP(respList, reqList)

	var result struct {
		Products []struct {
			ID int `json:"id"`
		} `json:"products"`
	}
	json.Unmarshal(respList.Body.Bytes(), &result)
	if len(result.Products) == 0 {
		t.Fatalf("Nenhum produto retornado para atualizar")
	}
	id := result.Products[len(result.Products)-1].ID

	// Atualiza produto usando ProductToAdd como base
	updatedProduct := product
	updatedProduct.Name = "Produto Atualizado"
	updatedProduct.Description = "Descrição Atualizada"
	updatedProduct.Price = 199.99
	updatedProduct.Stock = 15

	updateBody, err := json.Marshal(updatedProduct)
	if err != nil {
		t.Fatalf("Erro ao converter produto atualizado para JSON: %v", err)
	}

	reqUpdate, _ := http.NewRequest("PUT", "/products/"+strconv.Itoa(id), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate := httptest.NewRecorder()
	r.ServeHTTP(respUpdate, reqUpdate)

	if respUpdate.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", respUpdate.Code)
	}
}

func TestDeleteProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/products", CreateProductHandler)
	r.DELETE("/products/:id", DeleteProductHandler)
	r.GET("/products", ListProductsHandler)

	// Cria produto usando ProductToAdd
	product := models.ProductToAdd
	body, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Erro ao converter ProductToAdd para JSON: %v", err)
	}

	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// Lista produtos
	reqList, _ := http.NewRequest("GET", "/products", nil)
	respList := httptest.NewRecorder()
	r.ServeHTTP(respList, reqList)

	var result struct {
		Products []struct {
			ID int `json:"id"`
		} `json:"products"`
	}
	json.Unmarshal(respList.Body.Bytes(), &result)
	if len(result.Products) == 0 {
		t.Fatalf("Nenhum produto retornado para deletar")
	}
	id := result.Products[len(result.Products)-1].ID

	// Deleta produto
	reqDel, _ := http.NewRequest("DELETE", "/products/"+strconv.Itoa(id), nil)
	delResp := httptest.NewRecorder()
	r.ServeHTTP(delResp, reqDel)

	if delResp.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", delResp.Code)
	}
}
