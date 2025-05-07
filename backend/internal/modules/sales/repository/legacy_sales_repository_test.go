package repository

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

func TestCreateSale(t *testing.T) {
	s := models.Sale{
		Product:  "Teste Produto",
		Quantity: 5,
		Price:    99.99,
		Customer: "cliente@teste.com",
	}

	created, err := CreateSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda: %v", err)
	}

	if created.ID == 0 {
		t.Error("Esperava um ID gerado para a venda")
	}
}

func TestGetAllSales(t *testing.T) {
	sales, err := GetAllSales()
	if err != nil {
		t.Fatalf("Erro ao buscar vendas: %v", err)
	}
	t.Logf("Total de vendas encontradas: %d", len(sales))
}

func TestGetSaleByID(t *testing.T) {
	s := models.Sale{
		Product:  "Produto para Teste GetByID",
		Quantity: 3,
		Price:    49.99,
		Customer: "getbyid@teste.com",
	}

	created, err := CreateSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda para teste: %v", err)
	}

	retrieved, err := GetSaleByID(created.ID)
	if err != nil {
		t.Fatalf("Erro ao buscar venda por ID: %v", err)
	}

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

	nonExistingID := 99999
	_, err = GetSaleByID(nonExistingID)
	if err == nil {
		t.Error("Esperava erro ao buscar venda inexistente, mas não ocorreu")
	}

	err = DeleteSale(created.ID)
	if err != nil {
		t.Logf("Aviso: Não foi possível limpar a venda de teste: %v", err)
	}
}

func TestUpdateSale(t *testing.T) {
	s := models.Sale{
		Product:  "Produto Antigo",
		Quantity: 1,
		Price:    10.0,
		Customer: "antigo@cliente.com",
	}
	created, err := CreateSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda para update: %v", err)
	}

	updated := models.Sale{
		Product:  "Produto Atualizado",
		Quantity: 10,
		Price:    55.0,
		Customer: "novo@cliente.com",
	}
	result, err := UpdateSale(created.ID, updated)
	if err != nil {
		t.Fatalf("Erro ao atualizar venda: %v", err)
	}

	if result.Product != updated.Product || result.Quantity != updated.Quantity {
		t.Error("Dados da venda não foram atualizados corretamente")
	}
}

func TestDeleteSale(t *testing.T) {
	s := models.Sale{
		Product:  "Produto para Deletar",
		Quantity: 2,
		Price:    20.0,
		Customer: "deletar@cliente.com",
	}
	created, err := CreateSale(s)
	if err != nil {
		t.Fatalf("Erro ao criar venda para deletar: %v", err)
	}

	err = DeleteSale(created.ID)
	if err != nil {
		t.Errorf("Erro ao deletar venda: %v", err)
	}

	// Confirma se a venda foi removida tentando deletar novamente
	err = DeleteSale(created.ID)
	if err == nil {
		t.Error("Esperava erro ao deletar venda inexistente, mas não ocorreu")
	}
}
