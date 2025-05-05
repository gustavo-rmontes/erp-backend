package service

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

func TestCreateAndListRentals(t *testing.T) {
	r := models.Rental{
		ClientName:  "Service Cliente",
		Equipment:   "Tablet",
		StartDate:   "2025-04-01",
		EndDate:     "2025-12-01",
		Price:       450,
		BillingType: "mensal",
	}
	if err := CreateRental(r); err != nil {
		t.Fatalf("Erro ao criar locação: %v", err)
	}
	list, err := ListRentals()
	if err != nil || len(list) == 0 {
		t.Errorf("Erro ao listar locações ou lista vazia")
	}
}

func TestUpdateAndDeleteRental(t *testing.T) {
	r := models.Rental{
		ClientName:  "Para Atualizar",
		Equipment:   "Impressora",
		StartDate:   "2025-06-01",
		EndDate:     "2026-06-01",
		Price:       200,
		BillingType: "anual",
	}
	CreateRental(r)
	rentals, _ := ListRentals()
	id := rentals[len(rentals)-1].ID

	rUpdated := models.Rental{
		ClientName:  "Atualizado",
		Equipment:   "Impressora Pro",
		StartDate:   "2025-07-01",
		EndDate:     "2026-07-01",
		Price:       300,
		BillingType: "mensal",
	}
	if err := UpdateRental(id, rUpdated); err != nil {
		t.Errorf("Erro ao atualizar locação: %v", err)
	}

	if err := RemoveRental(id); err != nil {
		t.Errorf("Erro ao deletar locação: %v", err)
	}
}
