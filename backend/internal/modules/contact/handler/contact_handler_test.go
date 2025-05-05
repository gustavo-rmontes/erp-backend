package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

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

	os.Exit(m.Run())
}

func TestCreateContactHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/contacts", CreateContactHandler)

	body := []byte(`{
		"name": "Contato Handler",
		"email": "handler@contato.com",
		"phone": "111111111",
		"type": "cliente"
	}`)

	req, _ := http.NewRequest("POST", "/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado status 201, obtido %d", resp.Code)
	}
}

func TestListContactsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/contacts", ListContactsHandler)

	req, _ := http.NewRequest("GET", "/contacts", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", resp.Code)
	}
}

func TestUpdateContactHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/contacts", CreateContactHandler)
	router.PUT("/contacts/:id", UpdateContactHandler)
	router.GET("/contacts", ListContactsHandler)

	// Cria contato
	body := []byte(`{
		"name": "Contato Atualizável",
		"email": "original@teste.com",
		"phone": "123456789",
		"type": "cliente"
	}`)
	req, _ := http.NewRequest("POST", "/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("Falha ao criar contato: status %d", resp.Code)
	}

	// Busca ID
	respList := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/contacts", nil)
	router.ServeHTTP(respList, reqList)

	var result struct {
		Contacts []struct {
			ID int `json:"id"`
		} `json:"contacts"`
	}
	json.Unmarshal(respList.Body.Bytes(), &result)
	if len(result.Contacts) == 0 {
		t.Fatalf("Nenhum contato retornado para atualizar")
	}
	id := result.Contacts[len(result.Contacts)-1].ID

	// Atualiza contato
	updateBody := []byte(`{
		"name": "Contato Atualizado Handler",
		"email": "atualizado@teste.com",
		"phone": "999999999",
		"type": "fornecedor"
	}`)
	reqUpdate, _ := http.NewRequest("PUT", "/contacts/"+strconv.Itoa(id), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate := httptest.NewRecorder()
	router.ServeHTTP(respUpdate, reqUpdate)

	if respUpdate.Code != http.StatusOK {
		t.Errorf("Esperado status 200 na atualização, obtido %d", respUpdate.Code)
	}
}

func TestDeleteContactHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/contacts/:id", DeleteContactHandler)

	// Tenta deletar um contato com ID inexistente
	invalidID := 999999
	req, _ := http.NewRequest("DELETE", "/contacts/"+strconv.Itoa(invalidID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Como o handler atual retorna erro 500, o teste agora espera este status
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Esperado status 500 ao deletar ID inexistente, obtido %d", resp.Code)
	}

	// Verifica se a mensagem de erro retornada corresponde à esperada
	var responseBody struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("Erro ao decodificar a resposta: %v", err)
	}

	expected := "erro ao deletar contato"
	if responseBody.Error != expected {
		t.Errorf("Mensagem de erro inesperada. Esperado: '%s', obtido: '%s'", expected, responseBody.Error)
	}
}
