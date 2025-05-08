package service

import (
	"ERP-ONSMART/backend/internal/modules/sales/models"
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

func TestAddSale(t *testing.T) {
	s := models.Sale{
		Product:  "Produto Teste Service",
		Quantity: 2,
		Price:    29.99,
		Customer: "cliente@service.com",
	}

	created, err := AddSale(s)
	if err != nil {
		t.Fatalf("Erro ao adicionar venda: %v", err)
	}

	if created.ID == 0 {
		t.Error("Esperava um ID gerado para a venda")
	}
}

func TestListSales(t *testing.T) {
	sales, err := ListSales()
	if err != nil {
		t.Fatalf("Erro ao listar vendas: %v", err)
	}

	t.Logf("%d vendas listadas com sucesso", len(sales))
}

func TestGetSale(t *testing.T) {
	s := models.Sale{
		Product:  "Produto para Teste GetSale Service",
		Quantity: 4,
		Price:    59.99,
		Customer: "getsale@service.com",
	}

	created, err := AddSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda para teste: %v", err)
	}

	retrieved, err := GetSale(created.ID)
	if err != nil {
		t.Fatalf("Erro ao obter venda por ID: %v", err)
	}

	// Verify all fields match
	if retrieved.ID != created.ID {
		t.Errorf("ID incorreto, esperado %d, obtido %d", created.ID, retrieved.ID)
	}
	if retrieved.Product != created.Product {
		t.Errorf("Produto incorreto, esperado %s, obtido %s", created.Product, retrieved.Product)
	}
	if retrieved.Quantity != created.Quantity {
		t.Errorf("Quantidade incorreta, esperada %d, obtida %d", created.Quantity, retrieved.Quantity)
	}
	if retrieved.Price != created.Price {
		t.Errorf("Preço incorreto, esperado %.2f, obtido %.2f", created.Price, retrieved.Price)
	}
	if retrieved.Customer != created.Customer {
		t.Errorf("Cliente incorreto, esperado %s, obtido %s", created.Customer, retrieved.Customer)
	}

	// Test trying to retrieve a non-existent sale
	nonExistingID := 99999
	_, err = GetSale(nonExistingID)
	if err == nil {
		t.Error("Esperava erro ao buscar venda inexistente, mas não ocorreu")
	}

	err = RemoveSale(created.ID)
	if err != nil {
		t.Logf("Aviso: Não foi possível limpar a venda de teste: %v", err)
	}
}

func TestModifySale(t *testing.T) {
	original := models.Sale{
		Product:  "Produto Original",
		Quantity: 1,
		Price:    10.0,
		Customer: "cliente@original.com",
	}

	created, err := AddSale(original)
	if err != nil {
		t.Fatalf("Erro ao criar venda para update: %v", err)
	}

	update := models.Sale{
		Product:  "Produto Atualizado",
		Quantity: 5,
		Price:    50.0,
		Customer: "cliente@atualizado.com",
	}

	updated, err := ModifySale(created.ID, update)
	if err != nil {
		t.Fatalf("Erro ao atualizar venda: %v", err)
	}

	if updated.Product != update.Product || updated.Quantity != update.Quantity {
		t.Error("A venda não foi atualizada corretamente")
	}
}

func TestRemoveSale(t *testing.T) {
	s := models.Sale{
		Product:  "Produto Remoção",
		Quantity: 3,
		Price:    35.0,
		Customer: "cliente@remocao.com",
	}

	created, err := AddSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda para remoção: %v", err)
	}

	err = RemoveSale(created.ID)
	if err != nil {
		t.Fatalf("Erro ao remover venda: %v", err)
	}

	err = RemoveSale(created.ID)
	if err == nil {
		t.Error("Esperava erro ao tentar remover venda já deletada")
	}
}
