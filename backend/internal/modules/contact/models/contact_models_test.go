package models

import "testing"

func TestContactModel(t *testing.T) {
	c := Contact{
		Name:  "Empresa X",
		Email: "contato@empresa.com",
		Phone: "11999999999",
		Type:  "cliente",
	}
	if c.Name == "" || c.Email == "" || c.Type == "" {
		t.Errorf("Campos obrigatórios não preenchidos corretamente")
	}
}
