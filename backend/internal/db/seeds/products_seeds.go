package seeds

import (
	"database/sql"
	"fmt"
	"log"

	"ERP-ONSMART/backend/internal/modules/products/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedProducts gera produtos fictícios
func SeedProducts(db *sql.DB, count int) error {
	log.Printf("[seeds:products] Iniciando geração de %d produtos...", count)

	// Verificar se a tabela products existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'products')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:products] Erro ao verificar existência da tabela 'products': %w", err)
	}

	if !exists {
		log.Printf("[seeds:products] Tabela 'products' não existe. Seed de produtos será ignorado.")
		return nil
	}

	// Prepare statement sem a coluna created_at
	stmt, err := db.Prepare(`
        INSERT INTO products 
        (name, description, price, stock) 
        VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:products] Erro ao preparar inserção de produtos: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:products] Inserção preparada com sucesso.")

	for i := 0; i < count; i++ {
		// Gerar dados fictícios para o produto
		product := models.Product{
			Name:        gofakeit.ProductName(),
			Description: gofakeit.Paragraph(1, 3, 5, " "),
			Price:       gofakeit.Price(10, 1000),
			Stock:       gofakeit.Number(1, 100),
		}

		_, err := stmt.Exec(
			product.Name,
			product.Description,
			product.Price,
			product.Stock,
		)

		if err != nil {
			return fmt.Errorf("[seeds:products] Erro ao inserir produto #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:products] Geração de produtos concluída com sucesso.")
	return nil
}
