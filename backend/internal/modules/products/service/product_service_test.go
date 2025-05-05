package service

import (
	models "ERP-ONSMART/backend/internal/modules/products/models"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestCreateProduct(t *testing.T) {
	p := models.ProductToAdd

	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto via service: %v", err)
	} else {
		t.Logf("Produto criado com sucesso: %+v", p)
	}
	products, err := ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos após criação: %v", err)
	}
	// Assume que o último produto é o recém-criado
	createdProduct := products[len(products)-1]
	DeleteProduct(createdProduct.ID)
}

func TestListProducts(t *testing.T) {
	products, err := ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos via service: %v", err)
	}
	if len(products) == 0 {
		t.Fatalf("Nenhum produto encontrado via service")
	}
}

func TestListProductByID(t *testing.T) {
	// Cria um produto para o teste
	p := models.ProductToAdd
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para teste de ListProductByID: %v", err)
	}

	// Obtém o ID do produto recém-criado
	products, err := ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos para obter ID: %v", err)
	}
	id := products[len(products)-1].ID

	// Busca o produto pelo ID
	product, err := ListProductByID(id)
	if err != nil {
		t.Fatalf("Erro ao buscar produto por ID via service: %v", err)
	}

	// Verifica se o produto retornado é o esperado
	if product.ID != id {
		t.Fatalf("Produto retornado não corresponde ao esperado. Esperado ID: %d, Obtido ID: %d", id, product.ID)
	}

	t.Logf("Produto encontrado com sucesso: %+v", product)

	// Limpa o banco de dados
	if err := DeleteProduct(id); err != nil {
		t.Fatalf("Erro ao deletar produto após teste de ListProductByID: %v", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	p := models.ProductToAdd

	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para update: %v", err)
	}

	// Busca a lista de produtos e obtém o ID do último inserido
	products, err := ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos para update: %v", err)
	}
	id := products[len(products)-1].ID

	// Define os novos dados para atualização
	updated := models.ProductToUpdate
	if err := UpdateProduct(id, updated); err != nil {
		t.Fatalf("Erro ao atualizar produto via service: %v", err)
	}

	// Lista novamente para verificar se a atualização ocorreu corretamente
	products, err = ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos após update: %v", err)
	}
	// Verifica o produto atualizado (assumindo que o produto atualizado é o último da lista)
	lastProduct := products[len(products)-1]
	if lastProduct.Name != updated.Name ||
		lastProduct.DetailedName != updated.DetailedName ||
		lastProduct.Description != updated.Description ||
		lastProduct.Status != updated.Status ||
		lastProduct.SKU != updated.SKU ||
		lastProduct.Barcode != updated.Barcode ||
		lastProduct.ExternalID != updated.ExternalID ||
		lastProduct.Coin != updated.Coin ||
		lastProduct.Price != updated.Price ||
		lastProduct.SalesPrice != updated.SalesPrice ||
		lastProduct.CostPrice != updated.CostPrice ||
		lastProduct.Stock != updated.Stock ||
		lastProduct.Type != updated.Type ||
		lastProduct.ProductGroup != updated.ProductGroup ||
		lastProduct.ProductCategory != updated.ProductCategory ||
		lastProduct.ProductSubcategory != updated.ProductSubcategory ||
		len(lastProduct.Tags) != len(updated.Tags) ||
		lastProduct.Manufacturer != updated.Manufacturer ||
		lastProduct.ManufacturerCode != updated.ManufacturerCode ||
		lastProduct.NCM != updated.NCM ||
		lastProduct.CEST != updated.CEST ||
		lastProduct.CNAE != updated.CNAE ||
		lastProduct.Origin != updated.Origin ||
		len(lastProduct.Images) != len(updated.Images) ||
		len(lastProduct.Documents) != len(updated.Documents) {
		t.Errorf("\nProduto não foi atualizado corretamente. \nEsperado: %+v \nObtido: %+v\n", updated, lastProduct)
	} else {
		t.Logf("\nProduto atualizado com sucesso: %+v", lastProduct)
		DeleteProduct(id)
	}
}

func TestDeleteProduct(t *testing.T) {
	// Cria um produto para deletar
	p := models.ProductToAdd
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para deleção via service: %v", err)
	}

	// Lista produtos para obter o ID do produto recém-criado
	products, err := ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos após criação para deleção: %v", err)
	}
	id := products[len(products)-1].ID

	// Deleta o produto via service
	if err := DeleteProduct(id); err != nil {
		t.Fatalf("Erro ao deletar produto via service: %v", err)
	}

	// Verifica se o produto foi realmente removido
	products, err = ListProducts()
	if err != nil {
		t.Fatalf("Erro ao listar produtos após deleção via service: %v", err)
	}
	for _, prod := range products {
		if prod.ID == id {
			t.Errorf("Produto com ID %d ainda existe após deleção", id)
		}
	}
}
