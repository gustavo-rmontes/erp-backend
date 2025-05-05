package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"ERP-ONSMART/backend/internal/modules/accounting/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedTransactions gera transações financeiras fictícias
func SeedTransactions(db *sql.DB, count int) error {
	log.Printf("[seeds:transactions] Iniciando geração de %d transações financeiras...", count)

	// Verificar se a tabela acc_transaction existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'acc_transaction')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:transactions] Erro ao verificar existência da tabela 'acc_transaction': %w", err)
	}

	if !exists {
		log.Printf("[seeds:transactions] Tabela 'acc_transaction' não existe. Seed de transações será ignorado.")
		return nil
	}

	// Preparar statement para inserção
	stmt, err := db.Prepare(`
        INSERT INTO acc_transaction 
        (description, amount, date) 
        VALUES ($1, $2, $3)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:transactions] Erro ao preparar inserção de transações: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:transactions] Inserção preparada com sucesso.")

	// Descrições possíveis para transações
	descriptions := []string{
		"Pagamento de fornecedor",
		"Recebimento de cliente",
		"Despesa operacional",
		"Folha de pagamento",
		"Imposto",
		"Investimento",
		"Manutenção",
		"Aluguel",
	}

	// Período de datas: últimos 12 meses
	startDate := time.Now().AddDate(-1, 0, 0)
	endDate := time.Now()

	for i := 0; i < count; i++ {
		// Gera uma data aleatória nos últimos 12 meses
		transactionDate := gofakeit.DateRange(startDate, endDate)

		// Formata no padrão ISO (YYYY-MM-DD) para inserção no PostgreSQL
		formattedDateForDB := transactionDate.Format("2006-01-02")

		// Aleatoriamente positivo ou negativo (receita ou despesa)
		var amount float64
		if gofakeit.Bool() {
			amount = gofakeit.Price(100, 10000) // Receita
		} else {
			amount = -gofakeit.Price(100, 5000) // Despesa (valor negativo)
		}

		// Cria a transação usando o modelo
		transaction := models.Transaction{
			Description: descriptions[gofakeit.Number(0, len(descriptions)-1)],
			Amount:      amount,
			Date:        formattedDateForDB,
		}

		// Insere no banco usando o formato que o PostgreSQL espera
		_, err := stmt.Exec(
			transaction.Description,
			transaction.Amount,
			transaction.Date,
		)

		if err != nil {
			return fmt.Errorf("[seeds:transactions] Erro ao inserir transação #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:transactions] Geração de transações concluída com sucesso.")
	return nil
}
