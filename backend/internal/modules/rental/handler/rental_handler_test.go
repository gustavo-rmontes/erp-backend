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

func TestCreateRentalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/rentals", CreateRentalHandler)

	body := []byte(`{
		"client_name": "Locadora ABC",
		"equipment": "Servidor Dell",
		"start_date": "2025-04-01",
		"end_date": "2025-07-01",
		"price": 3000.75,
		"billing_type": "mensal"
	}`)

	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Esperado 201, obtido %d", resp.Code)
	}
}

func TestListRentalsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/rentals", ListRentalsHandler)

	req, _ := http.NewRequest("GET", "/rentals", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", resp.Code)
	}
}

func TestUpdateRentalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/rentals", CreateRentalHandler)
	r.PUT("/rentals/:id", UpdateRentalHandler)
	r.GET("/rentals", ListRentalsHandler)

	createBody := []byte(`{
		"client_name": "Cliente Update",
		"equipment": "Firewall",
		"start_date": "2025-05-01",
		"end_date": "2025-12-01",
		"price": 4200,
		"billing_type": "anual"
	}`)
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	reqList, _ := http.NewRequest("GET", "/rentals", nil)
	respList := httptest.NewRecorder()
	r.ServeHTTP(respList, reqList)

	var result struct {
		Rentals []struct {
			ID int `json:"id"`
		} `json:"rentals"`
	}
	json.Unmarshal(respList.Body.Bytes(), &result)
	if len(result.Rentals) == 0 {
		t.Fatalf("Nenhuma locação retornada para atualizar")
	}
	id := result.Rentals[len(result.Rentals)-1].ID

	updateBody := []byte(`{
		"client_name": "Cliente Atualizado",
		"equipment": "Servidor HP",
		"start_date": "2025-06-01",
		"end_date": "2025-12-01",
		"price": 4300,
		"billing_type": "mensal"
	}`)
	reqUpdate, _ := http.NewRequest("PUT", "/rentals/"+strconv.Itoa(id), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate := httptest.NewRecorder()
	r.ServeHTTP(respUpdate, reqUpdate)

	if respUpdate.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", respUpdate.Code)
	}
}

func TestDeleteRentalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/rentals", CreateRentalHandler)
	r.DELETE("/rentals/:id", DeleteRentalHandler)
	r.GET("/rentals", ListRentalsHandler)

	createBody := []byte(`{
		"client_name": "Cliente Delete",
		"equipment": "Switch",
		"start_date": "2025-04-01",
		"end_date": "2025-07-01",
		"price": 2000,
		"billing_type": "trimestral"
	}`)
	req, _ := http.NewRequest("POST", "/rentals", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	reqList, _ := http.NewRequest("GET", "/rentals", nil)
	respList := httptest.NewRecorder()
	r.ServeHTTP(respList, reqList)

	var result struct {
		Rentals []struct {
			ID int `json:"id"`
		} `json:"rentals"`
	}
	json.Unmarshal(respList.Body.Bytes(), &result)
	if len(result.Rentals) == 0 {
		t.Fatalf("Nenhuma locação retornada para deletar")
	}
	id := result.Rentals[len(result.Rentals)-1].ID

	reqDel, _ := http.NewRequest("DELETE", "/rentals/"+strconv.Itoa(id), nil)
	delResp := httptest.NewRecorder()
	r.ServeHTTP(delResp, reqDel)

	if delResp.Code != http.StatusOK {
		t.Errorf("Esperado 200, obtido %d", delResp.Code)
	}
}
