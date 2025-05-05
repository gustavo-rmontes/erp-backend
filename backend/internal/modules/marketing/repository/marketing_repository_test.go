package repository

import (
	"ERP-ONSMART/backend/internal/modules/marketing/models"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env") // ← Caminho ajustado
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestCreateCampaign(t *testing.T) {
	camp := models.Campaign{
		Title:       "Promoção",
		Description: "Descontos",
		Budget:      1000,
		StartDate:   "2023-01-01",
		EndDate:     "2023-01-31",
	}

	created, err := CreateCampaign(camp)
	if err != nil {
		t.Fatalf("Erro ao criar campanha: %v", err)
	}

	if created.ID == 0 {
		t.Errorf("Esperava ID válido, obtido: %d", created.ID)
	}
}

func TestGetCampaign(t *testing.T) {
	camps, err := GetAllCampaigns()
	if err != nil {
		t.Fatalf("Erro ao buscar campanhas: %v", err)
	}

	if len(camps) == 0 {
		t.Error("Nenhuma campanha encontrada")
	}
}

func TestUpdateCampaign(t *testing.T) {
	camp := models.Campaign{
		Title:       "Atualização",
		Description: "Campanha antiga",
		Budget:      500,
		StartDate:   "2023-02-01",
		EndDate:     "2023-02-28",
	}

	created, err := CreateCampaign(camp)
	if err != nil {
		t.Fatalf("Erro ao criar campanha: %v", err)
	}

	update := models.Campaign{
		Title:       "Atualizado",
		Description: "Descrição atualizada",
		Budget:      750,
		StartDate:   "2023-02-10",
		EndDate:     "2023-03-10",
	}

	updated, err := UpdateCampaign(created.ID, update)
	if err != nil {
		t.Fatalf("Erro ao atualizar campanha: %v", err)
	}

	if updated.Title != update.Title {
		t.Errorf("Esperava título '%s', obteve '%s'", update.Title, updated.Title)
	}
}

func TestDeleteCampaign(t *testing.T) {
	camp := models.Campaign{
		Title:       "Para deletar",
		Description: "Será removida",
		Budget:      300,
		StartDate:   "2023-04-01",
		EndDate:     "2023-04-30",
	}

	created, err := CreateCampaign(camp)
	if err != nil {
		t.Fatalf("Erro ao criar campanha para deletar: %v", err)
	}

	// Deleta a campanha
	err = DeleteCampaign(created.ID)
	if err != nil {
		t.Errorf("Erro ao deletar campanha: %v", err)
	}

	// Tenta deletar novamente: deve retornar erro informando que a campanha não foi encontrada
	err = DeleteCampaign(created.ID)
	if err == nil {
		t.Errorf("Esperava erro ao deletar campanha inexistente, mas não houve erro")
	} else {
		expected := fmt.Sprintf("Campanha com ID %d não encontrado", created.ID)
		if err.Error() != expected {
			t.Errorf("Esperava mensagem de erro '%s', obteve: '%v'", expected, err.Error())
		}
	}
}
