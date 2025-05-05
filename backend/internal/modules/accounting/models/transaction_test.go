package models

import "testing"

func TestTransactionModel(t *testing.T) {
	trans := Transaction{ID: 1, Description: "Teste", Amount: 100.0, Date: "01/01/2023"}
	if trans.ID != 1 {
		t.Errorf("Esperado ID 1, obtido %d", trans.ID)
	}
}
