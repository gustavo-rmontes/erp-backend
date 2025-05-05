package service

import (
	"ERP-ONSMART/backend/internal/modules/marketing/models"
	"ERP-ONSMART/backend/internal/modules/marketing/repository"
	"fmt"
	"time"
)

func ListCampaigns() ([]models.Campaign, error) {
	return repository.GetAllCampaigns()
}

func AddCampaign(c models.Campaign) (models.Campaign, error) {
	// Converte datas de dd/mm/yyyy → yyyy-mm-dd
	layoutBR := "02/01/2006"
	start, err := time.Parse(layoutBR, c.StartDate)
	if err != nil {
		return models.Campaign{}, fmt.Errorf("data inicial inválida: %w", err)
	}
	end, err := time.Parse(layoutBR, c.EndDate)
	if err != nil {
		return models.Campaign{}, fmt.Errorf("data final inválida: %w", err)
	}

	c.StartDate = start.Format("2006-01-02")
	c.EndDate = end.Format("2006-01-02")

	return repository.CreateCampaign(c)
}

func ModifyCampaign(id int, c models.Campaign) (models.Campaign, error) {
	layoutBR := "02/01/2006"
	start, err := time.Parse(layoutBR, c.StartDate)
	if err != nil {
		return models.Campaign{}, fmt.Errorf("data inicial inválida: %w", err)
	}
	end, err := time.Parse(layoutBR, c.EndDate)
	if err != nil {
		return models.Campaign{}, fmt.Errorf("data final inválida: %w", err)
	}

	c.StartDate = start.Format("2006-01-02")
	c.EndDate = end.Format("2006-01-02")

	return repository.UpdateCampaign(id, c)
}

func RemoveCampaign(id int) error {
	return repository.DeleteCampaign(id)
}
