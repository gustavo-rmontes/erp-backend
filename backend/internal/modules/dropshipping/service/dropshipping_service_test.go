package service

import (
	"errors"
	"os"
	"testing"

	"ERP-ONSMART/backend/internal/modules/dropshipping/models"

	"github.com/spf13/viper"
)

// TestMain configura as variáveis de ambiente para os testes.
func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env") // Ajuste o caminho conforme necessário.
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestAddDropshipping testa a criação de uma transação de dropshipping.
// Se TotalPrice for zero, o service deve calcular TotalPrice = Price * Quantity.
func TestAddDropshipping(t *testing.T) {
	// Preparar registro de dropshipping (certifique-se de que os valores de ProductID, WarrantyID, etc. sejam válidos no seu ambiente).
	ds := models.Dropshipping{
		ProductID:  10,             // Substitua por um ID de produto existente
		WarrantyID: 20,             // Substitua por um ID de garantia existente
		Cliente:    "Test Cliente", // Campo do tipo string para contato
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 0, // Será calculado automaticamente
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	created, err := AddDropshipping(ds)
	if err != nil {
		t.Fatalf("AddDropshipping falhou: %v", err)
	}
	if created.ID == 0 {
		t.Errorf("ID não foi atribuído ao dropshipping inserido")
	}
	// Verifica se o TotalPrice foi calculado corretamente.
	expectedTotal := ds.Price * float64(ds.Quantity)
	if created.TotalPrice != expectedTotal {
		t.Errorf("TotalPrice incorreto. Esperado: %.2f, obtido: %.2f", expectedTotal, created.TotalPrice)
	}
}

// TestListDropshippings testa se a listagem retorna pelo menos um registro.
func TestListDropshippings(t *testing.T) {
	list, err := ListDropshippings()
	if err != nil {
		t.Fatalf("ListDropshippings falhou: %v", err)
	}
	if len(list) == 0 {
		t.Error("Nenhum dropshipping retornado na listagem")
	}
}

// TestModifyDropshipping testa a atualização de uma transação de dropshipping.
func TestModifyDropshipping(t *testing.T) {
	// Primeiro, cria um registro.
	ds := models.Dropshipping{
		ProductID:  10,
		WarrantyID: 20,
		Cliente:    "Test Cliente",
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 0,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	created, err := AddDropshipping(ds)
	if err != nil {
		t.Fatalf("Erro ao criar dropshipping para modificação: %v", err)
	}

	// Atualiza alguns campos.
	created.Price = 130.00
	created.Quantity = 3
	// O service recalcula o total automaticamente.
	updated, err := ModifyDropshipping(created.ID, created)
	if err != nil {
		t.Fatalf("ModifyDropshipping falhou: %v", err)
	}
	expectedTotal := created.Price * float64(created.Quantity)
	if updated.Price != 130.00 || updated.Quantity != 3 || updated.TotalPrice != expectedTotal {
		t.Errorf("Dropshipping não foi atualizado corretamente. Esperado: Price=130, Quantity=3, TotalPrice=%.2f; Obtido: Price=%.2f, Quantity=%d, TotalPrice=%.2f",
			expectedTotal, updated.Price, updated.Quantity, updated.TotalPrice)
	}
}

// TestRemoveDropshipping testa a remoção de uma transação de dropshipping.
func TestRemoveDropshipping(t *testing.T) {
	// Cria um registro de dropshipping.
	ds := models.Dropshipping{
		ProductID:  10,
		WarrantyID: 20,
		Cliente:    "Test Cliente",
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 240.00,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	created, err := AddDropshipping(ds)
	if err != nil {
		t.Fatalf("Erro ao criar dropshipping para remoção: %v", err)
	}
	// Remove o registro.
	err = RemoveDropshipping(created.ID)
	if err != nil {
		t.Fatalf("Erro ao remover dropshipping: %v", err)
	}
	// Tenta recuperar o registro removido.
	_, err = GetDropshipping(created.ID)
	if err == nil {
		t.Errorf("Dropshipping com ID %d não foi removido", created.ID)
	}
	// Opcional: verificar se o erro corresponde a "não encontrado".
	if !errors.Is(err, errors.New("registro de dropshipping não encontrado após inserção")) {
		// Aqui você pode ajustar a comparação conforme a mensagem de erro retornada.
		t.Logf("Erro esperado, mas verifique a mensagem: %v", err)
	}
}
