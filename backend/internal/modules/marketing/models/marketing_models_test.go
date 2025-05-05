package models

import "testing"

func TestCampaignModel(t *testing.T) {
	c := Campaign{ID: 1, Title: "Campanha Teste", Budget: 1000, StartDate: "01/01/2023", EndDate: "31/01/2023"}
	if c.ID != 1 {
		t.Errorf("Esperado ID 1, obtido %d", c.ID)
	}
}
