package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/products/models"
	"testing"

	"github.com/spf13/viper"
)

// Configura o Viper para carregar as variáveis de ambiente
func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}

	// Executa os testes
	m.Run()
}

// Testa a criação do produto
func TestCreateProduct(t *testing.T) {
	conn, err := db.OpenGormDB()
	if err != nil {
		t.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer func() {
		sqlDB, _ := conn.DB()
		sqlDB.Close()
	}()

	p := models.ProductToAdd

	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto: %v", err)
	} else {
		t.Logf("Produto criado com sucesso: %v", p)
		DeleteProductByID(p.ID)
	}
}

// Busca todos os produtos
func TestGetAllProducts(t *testing.T) {
	products, err := GetAllProducts()
	if err != nil {
		t.Fatalf("Erro ao buscar produtos: %v", err)
	}

	if len(products) == 0 {
		t.Fatalf("Nenhum produto encontrado")
	}
}

func TestGetProductByID(t *testing.T) {
	// Cria um produto para o teste
	p := models.ProductToAdd
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para busca: %v", err)
	}

	// Obtém o ID do produto criado
	products, _ := GetAllProducts()
	id := products[len(products)-1].ID

	// Busca o produto pelo ID
	product, err := GetProductByID(id)
	if err != nil {
		t.Fatalf("Erro ao buscar produto por ID: %v", err)
	}

	// Verifica se o produto retornado é o esperado
	if product.ID != id {
		t.Fatalf("ID do produto retornado não corresponde ao esperado. Esperado: %d, Obtido: %d", id, product.ID)
	}

	t.Logf("Produto encontrado com sucesso: %v", product)

	// Limpa o banco de dados
	DeleteProductByID(id)
}

// Atualiza o produto por ID
func TestUpdateProductByID(t *testing.T) {
	p := models.ProductToAdd
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para atualização: %v", err)
	}

	products, _ := GetAllProducts()
	id := products[len(products)-1].ID

	updated := models.ProductToUpdate
	if err := UpdateProductByID(id, updated); err != nil {
		t.Fatalf("Erro ao atualizar produto: %v", err)
	} else {
		t.Logf("Produto atualizado com sucesso: %v", updated)
	}
	DeleteProductByID(id)
}

// Deleta o produto por ID
func TestDeleteProductByID(t *testing.T) {
	p := models.ProductToAdd
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para deleção: %v", err)
	}

	products, _ := GetAllProducts()
	id := products[len(products)-1].ID

	if err := DeleteProductByID(id); err != nil {
		t.Fatalf("Erro ao deletar produto: %v", err)
	}
}
