package models

import "testing"

func TestRentalModel(t *testing.T) {
	r := Rental{
		ClientName:  "Empresa Teste",
		Equipment:   "Notebook",
		StartDate:   "2025-04-01",
		EndDate:     "2025-10-01",
		Price:       1500.50,
		BillingType: "mensal",
	}

	if r.ClientName == "" || r.Equipment == "" || r.BillingType == "" {
		t.Errorf("Campos obrigatórios não preenchidos")
	}
}
