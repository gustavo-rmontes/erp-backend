package repository

import (
	m "ERP-ONSMART/backend/internal/modules/dropshipping/models" // Models de dropshipping
	p "ERP-ONSMART/backend/internal/modules/products/models"     // Models de produtos
	ps "ERP-ONSMART/backend/internal/modules/products/service"   // Service de produtos
	"os"
	"testing"

	"github.com/spf13/viper"
)

// TestMain configura as variáveis de ambiente.
func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

// Helper para inserir um produto via service de produtos e retornar seu ID.
func createTestProduct(t *testing.T) int {
	product := p.Product{
		Name:        "Produto Dropshipping Teste",
		Description: "Produto inserido para teste de dropshipping",
		Price:       100.00,
		Stock:       50,
	}
	if err := ps.CreateProduct(&product); err != nil {
		t.Fatalf("Erro ao inserir produto: %v", err)
	}
	products, err := ps.ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao listar produtos: %v", err)
	}
	// Considera que o último produto inserido é o de teste.
	return products[len(products)-1].ID
}

// Helper para inserir uma garantia via service de produtos e retornar seu ID.
func createTestWarranty(t *testing.T, productID int) int {
	warranty := p.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          20.00,
	}
	if err := ps.CreateWarranty(warranty); err != nil {
		t.Fatalf("Erro ao inserir garantia: %v", err)
	}
	warranties, err := ps.ListWarranties()
	if err != nil || len(warranties) == 0 {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	return warranties[len(warranties)-1].ID
}

// Como o novo model de dropshipping utiliza um campo do tipo string para "contact" (Cliente),
// usamos um valor fixo para testes.
func testContact() string {
	return "Test Cliente"
}

// TestInsertDropshipping insere uma transação de dropshipping e garante que não haja erros.
func TestInsertDropshipping(t *testing.T) {
	productID := createTestProduct(t)
	warrantyID := createTestWarranty(t, productID)
	contact := testContact()

	ds := m.Dropshipping{
		ProductID:  productID,
		WarrantyID: warrantyID,
		Cliente:    contact,
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 240.00,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	if err := InsertDropshipping(ds); err != nil {
		t.Fatalf("Erro ao inserir dropshipping: %v", err)
	}
}

// TestGetDropshipping insere uma transação de dropshipping e utiliza o ID para buscá-la, validando a recuperação.
func TestGetDropshipping(t *testing.T) {
	productID := createTestProduct(t)
	warrantyID := createTestWarranty(t, productID)
	contact := testContact()

	ds := m.Dropshipping{
		ProductID:  productID,
		WarrantyID: warrantyID,
		Cliente:    contact,
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 240.00,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	if err := InsertDropshipping(ds); err != nil {
		t.Fatalf("Erro ao inserir dropshipping: %v", err)
	}

	list, err := GetAllDropshippings()
	if err != nil {
		t.Fatalf("Erro ao listar dropshippings: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("Lista de dropshippings vazia")
	}
	last := list[len(list)-1]

	retrieved, err := GetDropshippingByID(last.ID)
	if err != nil {
		t.Fatalf("Erro ao recuperar dropshipping por ID: %v", err)
	}
	if retrieved.ID != last.ID {
		t.Errorf("Dropshipping recuperado difere. Esperado ID %d, obtido %d", last.ID, retrieved.ID)
	}
}

// TestUpdateDropshipping insere um registro de dropshipping, atualiza alguns campos e valida a atualização.
func TestUpdateDropshipping(t *testing.T) {
	productID := createTestProduct(t)
	warrantyID := createTestWarranty(t, productID)
	contact := testContact()

	ds := m.Dropshipping{
		ProductID:  productID,
		WarrantyID: warrantyID,
		Cliente:    contact,
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 240.00,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	if err := InsertDropshipping(ds); err != nil {
		t.Fatalf("Erro ao inserir dropshipping: %v", err)
	}

	list, err := GetAllDropshippings()
	if err != nil || len(list) == 0 {
		t.Fatalf("Erro ao listar dropshippings: %v", err)
	}
	last := list[len(list)-1]

	// Atualiza campos
	last.Quantity = 3
	last.Price = 130.00
	last.TotalPrice = last.Price * float64(last.Quantity)
	last.UpdatedAt = "2025-04-10"

	if err := UpdateDropshippingByID(last.ID, last); err != nil {
		t.Fatalf("Erro ao atualizar dropshipping: %v", err)
	}

	updated, err := GetDropshippingByID(last.ID)
	if err != nil {
		t.Fatalf("Erro ao recuperar dropshipping atualizado: %v", err)
	}
	if updated.Quantity != 3 || updated.Price != 130.00 || updated.TotalPrice != 390.00 {
		t.Errorf("Dropshipping não foi atualizado corretamente: %+v", updated)
	}
}

// TestDeleteDropshipping insere um registro de dropshipping, o deleta e valida a remoção.
func TestDeleteDropshipping(t *testing.T) {
	productID := createTestProduct(t)
	warrantyID := createTestWarranty(t, productID)
	contact := testContact()

	ds := m.Dropshipping{
		ProductID:  productID,
		WarrantyID: warrantyID,
		Cliente:    contact,
		Price:      120.00,
		Quantity:   2,
		TotalPrice: 240.00,
		StartDate:  "2025-04-09",
		UpdatedAt:  "2025-04-09",
	}
	if err := InsertDropshipping(ds); err != nil {
		t.Fatalf("Erro ao inserir dropshipping: %v", err)
	}

	list, err := GetAllDropshippings()
	if err != nil || len(list) == 0 {
		t.Fatalf("Erro ao listar dropshippings: %v", err)
	}
	last := list[len(list)-1]

	if err := DeleteDropshippingByID(last.ID); err != nil {
		t.Fatalf("Erro ao deletar dropshipping: %v", err)
	}

	_, err = GetDropshippingByID(last.ID)
	if err == nil {
		t.Errorf("Dropshipping com ID %d não foi excluído", last.ID)
	}
}
