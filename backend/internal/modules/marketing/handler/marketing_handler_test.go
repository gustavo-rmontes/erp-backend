package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/marketing/models"
	"ERP-ONSMART/backend/internal/modules/marketing/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}

	// Inicializa o logger global
	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/campaigns", ListCampaignsHandler)
	r.POST("/campaigns", CreateCampaignHandler)
	r.PUT("/campaigns/:id", UpdateCampaignHandler)
	r.DELETE("/campaigns/:id", DeleteCampaignHandler)
	return r
}

func TestListCampaignsHandler(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/campaigns", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestAddCampaign(t *testing.T) {
	camp := models.Campaign{
		Title:       "Campanha Teste Add",
		Description: "Testando criação direta via service",
		Budget:      1000.0,
		StartDate:   "01/05/2025",
		EndDate:     "31/05/2025",
	}

	created, err := service.AddCampaign(camp)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID, "ID da campanha criada deve ser diferente de zero")

	// Verifica se os dados foram persistidos corretamente (opcional)
	assert.Equal(t, camp.Title, created.Title)
	assert.Equal(t, camp.Description, created.Description)
	assert.Equal(t, camp.Budget, created.Budget)
	assert.Equal(t, "2025-05-01", created.StartDate)
	assert.Equal(t, "2025-05-31", created.EndDate)
}

func TestUpdateCampaignHandler(t *testing.T) {
	router := setupRouter()

	// Cria campanha válida (com datas no formato BR, que o handler espera)
	camp := models.Campaign{
		Title:       "Atualizar Teste",
		Description: "Campanha update",
		Budget:      1500.0,
		StartDate:   "01/01/2025",
		EndDate:     "31/12/2025",
	}

	created, err := service.AddCampaign(camp)
	if err != nil {
		t.Fatalf("Erro ao criar campanha: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("ID inválido ao criar campanha")
	}

	// Corpo atualizado com campos válidos no formato BR
	updateBody := []byte(`{
		"title": "Campanha Atualizada",
		"description": "Atualização da campanha",
		"budget": 2500.0,
		"StartDate": "01/02/2025",
		"EndDate": "30/11/2025"
	}`)

	req, _ := http.NewRequest("PUT", "/campaigns/"+strconv.Itoa(created.ID), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUpdateCampaignHandler_NotFound(t *testing.T) {
	router := setupRouter()

	updateBody := []byte(`{
		"title": "Inexistente",
		"description": "Teste de campanha",
		"budget": 1000.0,
		"StartDate": "10/01/2025",
		"EndDate": "20/01/2025"
	}`)

	req, _ := http.NewRequest("PUT", "/campaigns/999999", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestDeleteCampaignHandler(t *testing.T) {
	router := setupRouter()

	camp := models.Campaign{
		Title:       "Campanha a Deletar",
		Description: "Será deletada",
		Budget:      500.0,
		StartDate:   "01/03/2025",
		EndDate:     "15/03/2025",
	}
	created, _ := service.AddCampaign(camp)

	req, _ := http.NewRequest("DELETE", "/campaigns/"+strconv.Itoa(created.ID), nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
