package models

import "testing"

func TestSaleModel_ValidFields(t *testing.T) {
	s := Sale{
		ID:       1,
		Product:  "Notebook",
		Quantity: 2,
		Price:    2500.00,
		Customer: "cliente@example.com",
	}

	if s.ID != 1 {
		t.Errorf("Esperado ID 1, obtido %d", s.ID)
	}
	if s.Product == "" {
		t.Error("Produto não deve estar vazio")
	}
	if s.Quantity <= 0 {
		t.Error("Quantidade deve ser maior que 0")
	}
	if s.Price <= 0 {
		t.Error("Preço deve ser maior que 0")
	}
	if !containsAtSymbol(s.Customer) {
		t.Errorf("Customer deve ser um e-mail válido, obtido: %s", s.Customer)
	}
}

func TestSaleModel_InvalidFields(t *testing.T) {
	s := Sale{
		Product:  "",
		Quantity: 0,
		Price:    0,
		Customer: "sem-arroba.com",
	}

	if s.Product != "" {
		t.Error("Produto deveria estar vazio para este teste")
	}
	if s.Quantity > 0 {
		t.Error("Quantidade deveria ser inválida (0 ou menor)")
	}
	if s.Price > 0 {
		t.Error("Preço deveria ser inválido (0 ou menor)")
	}
	if containsAtSymbol(s.Customer) {
		t.Errorf("Customer deveria ser inválido, mas recebeu: %s", s.Customer)
	}
}

func containsAtSymbol(email string) bool {
	for _, r := range email {
		if r == '@' {
			return true
		}
	}
	return false
}
