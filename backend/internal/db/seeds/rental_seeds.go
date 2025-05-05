package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"ERP-ONSMART/backend/internal/modules/rental/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedRentals gera aluguéis fictícios
func SeedRentals(db *sql.DB, count int) error {
	log.Printf("[seeds:rentals] Iniciando geração de %d aluguéis...", count)

	// Verificar se a tabela rentals existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'rentals')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:rentals] Erro ao verificar existência da tabela 'rentals': %w", err)
	}

	if !exists {
		log.Printf("[seeds:rentals] Tabela 'rentals' não existe. Seed de aluguéis será ignorado.")
		return nil
	}

	// Prepare statement com sintaxe PostgreSQL
	stmt, err := db.Prepare(`
        INSERT INTO rentals 
        (client_name, equipment, start_date, end_date, price, billing_type) 
        VALUES ($1, $2, $3, $4, $5, $6)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:rentals] Erro ao preparar inserção de aluguéis: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:rentals] Inserção preparada com sucesso.")

	// Tipos de cobrança possíveis
	billingTypes := []string{"mensal", "anual"}

	for i := 0; i < count; i++ {
		// Gerar dados fictícios para o aluguel
		rental := models.Rental{
			ClientName:  gofakeit.Name(),
			Equipment:   gofakeit.ProductName(),
			StartDate:   gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now()).Format("2006-01-02"),
			EndDate:     gofakeit.DateRange(time.Now(), time.Now().AddDate(0, 1, 0)).Format("2006-01-02"),
			Price:       gofakeit.Price(50, 500),
			BillingType: billingTypes[gofakeit.Number(0, 1)], // Alternando entre mensal e anual
		}

		_, err := stmt.Exec(
			rental.ClientName,
			rental.Equipment,
			rental.StartDate,
			rental.EndDate,
			rental.Price,
			rental.BillingType,
		)

		if err != nil {
			return fmt.Errorf("[seeds:rentals] Erro ao inserir aluguel #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:rentals] Geração de aluguéis concluída com sucesso.")
	return nil
}
