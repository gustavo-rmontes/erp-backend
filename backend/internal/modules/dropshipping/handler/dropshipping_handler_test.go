package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"ERP-ONSMART/backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Inicializa o logger com um "no-op" para evitar o nil pointer.
	logger.Logger = zap.NewNop()

	viper.SetConfigFile("../../../../../.env") // Ajuste o caminho conforme necessário.
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestCreateDropshippingHandler testa o endpoint POST para criar um dropshipping.
func TestCreateDropshippingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/dropshippings", CreateDropshippingHandler)

	// Payload: note que os campos obrigatórios devem ser preenchidos e os nomes dos campos
	// refletem o JSON tags do model Dropshipping.
	payload := `{
		"product_id": 1,
		"warranty_id": 1,
		"cliente": "Test Cliente",
		"price": 120.00,
		"quantity": 2,
		"total_price": 0,
		"start_date": "2025-04-09",
		"updated_at": "2025-04-09"
	}`
	req, err := http.NewRequest("POST", "/dropshippings", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado status 201, obtido %d, corpo: %s", resp.Code, resp.Body.String())
	}
}

// TestListDropshippingsHandler testa o endpoint GET para listar todos os dropshippings.
func TestListDropshippingsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/dropshippings", ListDropshippingsHandler)

	req, _ := http.NewRequest("GET", "/dropshippings", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", resp.Code)
	}
	// Opcional: decodificar a resposta para verificar se há o campo "data".
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Errorf("Erro ao decodificar resposta: %v", err)
	}
	if _, ok := result["data"]; !ok {
		t.Errorf("Resposta não contém o campo 'data'")
	}
}

// TestGetDropshippingHandler testa a recuperação de um dropshipping pelo ID.
// Este teste primeiro cria um dropshipping e, em seguida, tenta recuperá-lo.
func TestGetDropshippingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/dropshippings", CreateDropshippingHandler)
	router.GET("/dropshippings/:id", GetDropshippingHandler)

	// Cria um dropshipping para teste.
	payload := `{
		"product_id": 1,
		"warranty_id": 1,
		"cliente": "Test Cliente",
		"price": 120.00,
		"quantity": 2,
		"total_price": 0,
		"start_date": "2025-04-09",
		"updated_at": "2025-04-09"
	}`
	reqCreate, _ := http.NewRequest("POST", "/dropshippings", bytes.NewBuffer([]byte(payload)))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	if respCreate.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar dropshipping para teste: status %d", respCreate.Code)
	}

	// Decodifica a resposta para obter o ID.
	var created map[string]interface{}
	if err := json.Unmarshal(respCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("Erro ao decodificar resposta de criação: %v", err)
	}
	idFloat, ok := created["id"].(float64)
	if !ok {
		t.Fatalf("ID não retornado ou com tipo inválido")
	}
	id := int(idFloat)

	// Faz a requisição GET para recuperar o dropshipping pelo ID.
	reqGet, _ := http.NewRequest("GET", "/dropshippings/"+strconv.Itoa(id), nil)
	respGet := httptest.NewRecorder()
	router.ServeHTTP(respGet, reqGet)
	if respGet.Code != http.StatusOK {
		t.Errorf("Esperado status 200 ao recuperar dropshipping, obtido %d", respGet.Code)
	}
}

// TestUpdateDropshippingHandler testa a atualização de um dropshipping existente.
// O teste cria um registro, atualiza alguns campos e verifica a resposta.
func TestUpdateDropshippingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/dropshippings", CreateDropshippingHandler)
	router.PUT("/dropshippings/:id", UpdateDropshippingHandler)
	router.GET("/dropshippings/:id", GetDropshippingHandler)

	// Cria um registro para teste.
	payload := `{
		"product_id": 1,
		"warranty_id": 1,
		"cliente": "Test Cliente",
		"price": 120.00,
		"quantity": 2,
		"total_price": 0,
		"start_date": "2025-04-09",
		"updated_at": "2025-04-09"
	}`
	reqCreate, _ := http.NewRequest("POST", "/dropshippings", bytes.NewBuffer([]byte(payload)))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	if respCreate.Code != http.StatusCreated {
		t.Fatalf("Erro ao criar dropshipping: status %d", respCreate.Code)
	}

	// Extrai o ID do dropshipping criado.
	var created map[string]interface{}
	if err := json.Unmarshal(respCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}
	idFloat, ok := created["id"].(float64)
	if !ok {
		t.Fatalf("ID não retornado ou com tipo inválido")
	}
	id := int(idFloat)

	// Prepara o payload de atualização.
	updatePayload := `{
		"product_id": 1,
		"warranty_id": 1,
		"cliente": "Test Cliente Atualizado",
		"price": 130.00,
		"quantity": 3,
		"total_price": 0,
		"start_date": "2025-04-09",
		"updated_at": "2025-04-10"
	}`
	reqUpdate, _ := http.NewRequest("PUT", "/dropshippings/"+strconv.Itoa(id), bytes.NewBuffer([]byte(updatePayload)))
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate := httptest.NewRecorder()
	router.ServeHTTP(respUpdate, reqUpdate)
	if respUpdate.Code != http.StatusOK {
		t.Errorf("Esperado status 200 na atualização, obtido %d", respUpdate.Code)
	}

	// Verifica a atualização, fazendo um GET.
	reqGet, _ := http.NewRequest("GET", "/dropshippings/"+strconv.Itoa(id), nil)
	respGet := httptest.NewRecorder()
	router.ServeHTTP(respGet, reqGet)
	if respGet.Code != http.StatusOK {
		t.Errorf("Esperado status 200 ao recuperar dropshipping atualizado, obtido %d", respGet.Code)
	}
	var updated map[string]interface{}
	if err := json.Unmarshal(respGet.Body.Bytes(), &updated); err != nil {
		t.Fatalf("Erro ao decodificar resposta do GET: %v", err)
	}
	// Verifica se os campos foram atualizados.
	if p, ok := updated["price"].(float64); !ok || p != 130.00 {
		t.Errorf("Preço não foi atualizado corretamente; esperado 130.00, obtido %v", updated["price"])
	}
	if q, ok := updated["quantity"].(float64); !ok || q != 3 {
		t.Errorf("Quantidade não foi atualizada corretamente; esperado 3, obtido %v", updated["quantity"])
	}
	expectedTotal := 130.00 * 3
	if tot, ok := updated["total_price"].(float64); !ok || tot != expectedTotal {
		t.Errorf("TotalPrice não foi atualizado corretamente; esperado %.2f, obtido %v", expectedTotal, updated["total_price"])
	}
}

// TestDeleteDropshippingHandler_NotFound testa a remoção de um dropshipping inexistente.
func TestDeleteDropshippingHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/dropshippings/:id", DeleteDropshippingHandler)

	req, _ := http.NewRequest("DELETE", "/dropshippings/999999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Esperado status 404 ao deletar ID inexistente, obtido %d", resp.Code)
	}

	// Opcional: verifica se a mensagem de erro contém "não encontrado"
	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err == nil {
		if errMsg, ok := body["error"]; ok {
			if !strings.Contains(strings.ToLower(errMsg), "não encontrado") {
				t.Errorf("Mensagem de erro inesperada: %s", errMsg)
			}
		}
	}
}
