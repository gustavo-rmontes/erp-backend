package service

import (
	"ERP-ONSMART/backend/internal/modules/marketing/models"
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

func TestAddCampaign(t *testing.T) {
	c := models.Campaign{
		Title:       "Black Friday",
		Description: "Ofertas",
		Budget:      5000,
		StartDate:   "20/11/2023", // Formato ajustado: dd/mm/aaaa
		EndDate:     "25/11/2023", // Formato ajustado: dd/mm/aaaa
	}

	created, err := AddCampaign(c)
	if err != nil {
		t.Fatalf("Erro ao adicionar campanha: %v", err)
	}

	if created.ID == 0 {
		t.Errorf("Campanha não foi criada corretamente, ID retornado: %d", created.ID)
	}
}

func TestListCampaigns(t *testing.T) {
	camps, err := ListCampaigns()
	if err != nil {
		t.Fatalf("Erro ao listar campanhas: %v", err)
	}

	if len(camps) == 0 {
		t.Error("Esperava ao menos uma campanha cadastrada")
	}
}

func TestModifyCampaign(t *testing.T) {
	c := models.Campaign{
		Title:       "Pré-Alteração",
		Description: "Será atualizada",
		Budget:      1000,
		StartDate:   "01/10/2023", // Formato dd/mm/aaaa
		EndDate:     "15/10/2023", // Formato dd/mm/aaaa
	}
	created, err := AddCampaign(c)
	if err != nil {
		t.Fatalf("Erro ao adicionar campanha: %v", err)
	}

	updatedData := models.Campaign{
		Title:       "Pós-Alteração",
		Description: "Atualizada com sucesso",
		Budget:      1500,
		StartDate:   "05/10/2023", // Formato dd/mm/aaaa
		EndDate:     "20/10/2023", // Formato dd/mm/aaaa
	}

	updated, err := ModifyCampaign(created.ID, updatedData)
	if err != nil {
		t.Fatalf("Erro ao atualizar campanha: %v", err)
	}

	if updated.Title != updatedData.Title {
		t.Errorf("Esperava título '%s', obteve '%s'", updatedData.Title, updated.Title)
	}
}

func TestRemoveCampaign(t *testing.T) {
	c := models.Campaign{
		Title:       "Remoção",
		Description: "Campanha temporária",
		Budget:      500,
		StartDate:   "01/12/2023", // Formato dd/mm/aaaa
		EndDate:     "10/12/2023", // Formato dd/mm/aaaa
	}
	created, err := AddCampaign(c)
	if err != nil {
		t.Fatalf("Erro ao adicionar campanha para remoção: %v", err)
	}

	err = RemoveCampaign(created.ID)
	if err != nil {
		t.Fatalf("Erro ao remover campanha: %v", err)
	}
}
