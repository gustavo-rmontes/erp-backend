package seeds

import (
	"database/sql"
	"fmt"
	"log"

	"ERP-ONSMART/backend/internal/modules/sales/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedSales gera vendas fictícias
func SeedSales(db *sql.DB, count int) error {
	log.Printf("[seeds:sales] Iniciando geração de %d vendas...", count)

	// Verificar se existem produtos no banco
	var productCount int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	if err != nil {
		return fmt.Errorf("[seeds:sales] Erro ao verificar produtos existentes: %w", err)
	}

	// Preparar statement para inserção
	stmt, err := db.Prepare(`
        INSERT INTO sales 
        (product, quantity, price, customer) 
        VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:sales] Erro ao preparar inserção de vendas: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:sales] Inserção preparada com sucesso.")

	// Lista de produtos padrão caso não existam no banco
	defaultProducts := []string{
		"Notebook Dell",
		"Smartphone Samsung",
		"Monitor LG",
		"Impressora HP",
		"Teclado Mecânico",
		"Mouse Gamer",
		"Headset Bluetooth",
		"Tablet Apple",
	}

	// Obter nomes de produtos reais do banco, se existirem
	var productNames []string
	if productCount > 0 {
		rows, err := db.Query("SELECT name FROM products LIMIT 100")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err == nil {
					productNames = append(productNames, name)
				}
			}
		}
	}

	// Se não obtivemos produtos do banco, usar a lista padrão
	if len(productNames) == 0 {
		productNames = defaultProducts
	}

	for i := 0; i < count; i++ {
		// Escolhe um produto aleatório da lista
		productName := productNames[gofakeit.Number(0, len(productNames)-1)]

		// Gerar dados fictícios para a venda
		sale := models.Sale{
			Product:  productName,
			Quantity: gofakeit.Number(1, 10),
			Price:    gofakeit.Price(50, 5000),
			Customer: gofakeit.Email(), // O modelo requer um email para o cliente
		}

		_, err := stmt.Exec(
			sale.Product,
			sale.Quantity,
			sale.Price,
			sale.Customer,
		)

		if err != nil {
			return fmt.Errorf("[seeds:sales] Erro ao inserir venda #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:sales] Geração de vendas concluída com sucesso.")
	return nil
}
