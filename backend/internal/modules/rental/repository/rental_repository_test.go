package repository

import (
	"ERP-ONSMART/backend/internal/modules/rental/models"
	"os"
	"testing"

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

func TestInsertAndGetRentals(t *testing.T) {
	r := models.Rental{
		ClientName:  "Cliente Teste",
		Equipment:   "Router",
		StartDate:   "2025-04-01",
		EndDate:     "2025-05-01",
		Price:       100.0,
		BillingType: "mensal",
	}
	err := InsertRental(r)
	if err != nil {
		t.Fatalf("Erro ao inserir locação: %v", err)
	}

	rentals, err := GetAllRentals()
	if err != nil {
		t.Fatalf("Erro ao buscar locações: %v", err)
	}
	if len(rentals) == 0 {
		t.Error("Lista de locações vazia após inserção")
	}
}

func TestUpdateRentalByID(t *testing.T) {
	r := models.Rental{
		ClientName:  "Atualizar Teste",
		Equipment:   "Switch",
		StartDate:   "2025-04-10",
		EndDate:     "2025-06-10",
		Price:       200,
		BillingType: "anual",
	}
	InsertRental(r)
	rentals, _ := GetAllRentals()
	id := rentals[len(rentals)-1].ID

	updated := models.Rental{
		ClientName:  "Atualizado",
		Equipment:   "Switch X",
		StartDate:   "2025-05-01",
		EndDate:     "2025-07-01",
		Price:       250,
		BillingType: "trimestral",
	}
	err := UpdateRentalByID(id, updated)
	if err != nil {
		t.Errorf("Erro ao atualizar locação: %v", err)
	}
}

func TestDeleteRentalByID(t *testing.T) {
	r := models.Rental{
		ClientName:  "Deletar",
		Equipment:   "Servidor",
		StartDate:   "2025-01-01",
		EndDate:     "2025-03-01",
		Price:       999.99,
		BillingType: "mensal",
	}
	InsertRental(r)
	rentals, _ := GetAllRentals()
	id := rentals[len(rentals)-1].ID

	err := DeleteRentalByID(id)
	if err != nil {
		t.Errorf("Erro ao deletar locação: %v", err)
	}
}
