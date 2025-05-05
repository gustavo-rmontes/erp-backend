package service

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"testing"
)

// TestInsertWarranty insere uma garantia associada a um produto.
// Para associar a garantia, um produto é criado e inserido primeiro.
func TestInsertWarranty(t *testing.T) {
	// Cria um produto para associar à garantia
	product := models.Product{
		Name:        "Product for Warranty",
		Description: "Test product for warranty insertion",
		Price:       100.0,
		Stock:       10,
	}
	if err := CreateProduct(&product); err != nil {
		t.Fatalf("Erro ao criar produto: %v", err)
	}
	products, err := ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao listar produtos: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	warranty := models.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          15.0,
	}
	if err := CreateWarranty(warranty); err != nil {
		t.Fatalf("Erro ao criar garantia: %v", err)
	}
}

// TestGetWarranty recupera uma garantia pelo seu ID e compara seus dados.
func TestGetWarranty(t *testing.T) {
	// Lista todas as garantias cadastradas
	warranties, err := ListWarranties()
	if err != nil {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	if len(warranties) == 0 {
		t.Fatalf("Nenhuma garantia encontrada. Insira uma garantia antes de testar a obtenção.")
	}
	// Obtém a última garantia inserida para realizar o teste
	lastWarranty := warranties[len(warranties)-1]

	// Recupera a garantia pelo ID
	retrieved, err := GetWarrantyByID(lastWarranty.ID)
	if err != nil {
		t.Fatalf("Erro ao obter garantia: %v", err)
	}

	// Verifica se os dados recuperados estão corretos
	if retrieved.DurationMonths != lastWarranty.DurationMonths ||
		retrieved.Price != lastWarranty.Price ||
		retrieved.ProductID != lastWarranty.ProductID {
		t.Errorf("Dados da garantia incorretos. Esperado: %+v, obtido: %+v", lastWarranty, retrieved)
	}
}

func TestListWarranties(t *testing.T) {
	// Tenta listar todas as garantias existentes
	warranties, err := ListWarranties()
	if err != nil {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	if len(warranties) == 0 {
		t.Fatalf("Nenhuma garantia encontrada. Certifique-se de que pelo menos uma garantia foi inserida nos testes anteriores.")
	}
}

func TestUpdateWarranty(t *testing.T) {
	// Cria um produto para associar à garantia
	product := models.Product{
		Name:        "Product for Warranty Update",
		Description: "Product for warranty update testing",
		Price:       150.0,
		Stock:       5,
	}
	if err := CreateProduct(&product); err != nil {
		t.Fatalf("Erro ao criar produto: %v", err)
	}
	products, err := ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao listar produtos: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	warranty := models.Warranty{
		ProductID:      productID,
		DurationMonths: 6,
		Price:          20.0,
	}
	if err := CreateWarranty(warranty); err != nil {
		t.Fatalf("Erro ao criar garantia: %v", err)
	}
	warranties, err := ListWarranties()
	if err != nil || len(warranties) == 0 {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	createdWarranty := warranties[len(warranties)-1]

	// Atualiza os dados da garantia
	updatedWarranty := models.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          25.5,
	}
	if err := UpdateWarranty(createdWarranty.ID, updatedWarranty); err != nil {
		t.Fatalf("Erro ao atualizar garantia: %v", err)
	}

	// Recupera a garantia atualizada e verifica os dados
	retrieved, err := GetWarrantyByID(createdWarranty.ID)
	if err != nil {
		t.Fatalf("Erro ao buscar garantia atualizada: %v", err)
	}
	if retrieved.DurationMonths != updatedWarranty.DurationMonths || retrieved.Price != updatedWarranty.Price {
		t.Errorf("Garantia atualizada com dados incorretos. Esperado: %+v, obtido: %+v", updatedWarranty, retrieved)
	}
}

func TestDeleteWarranty(t *testing.T) {
	// Cria um produto para associar à garantia
	product := models.Product{
		Name:        "Product for Warranty Deletion",
		Description: "Product for warranty deletion testing",
		Price:       120.0,
		Stock:       8,
	}
	if err := CreateProduct(&product); err != nil {
		t.Fatalf("Erro ao criar produto: %v", err)
	}
	products, err := ListProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao listar produtos: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	warranty := models.Warranty{
		ProductID:      productID,
		DurationMonths: 9,
		Price:          18.0,
	}
	if err := CreateWarranty(warranty); err != nil {
		t.Fatalf("Erro ao criar garantia: %v", err)
	}
	warranties, err := ListWarranties()
	if err != nil || len(warranties) == 0 {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	createdWarranty := warranties[len(warranties)-1]

	// Deleta a garantia
	if err := DeleteWarranty(createdWarranty.ID); err != nil {
		t.Fatalf("Erro ao deletar garantia: %v", err)
	}

	// Tenta recuperar a garantia deletada; espera erro
	_, err = GetWarrantyByID(createdWarranty.ID)
	if err == nil {
		t.Errorf("Garantia com ID %d ainda existe após deleção", createdWarranty.ID)
	}
}
