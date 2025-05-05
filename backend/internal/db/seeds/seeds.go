package seeds

import (
	"database/sql"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedConfig para configurar a geração de dados
type SeedConfig struct {
	CustomersCount    int
	ProductsCount     int
	OrdersCount       int
	ContactsCount     int
	UsersCount        int
	TransactionsCount int
	CampaignsCount    int
	RentalsCount      int
	SalesCount        int
	Seed              int64 // Para reprodutibilidade
}

// ExecuteSeeds executa todos os seeds
func ExecuteSeeds(db *sql.DB, config SeedConfig) error {
	// Configura uma seed fixa para reprodutibilidade
	gofakeit.Seed(config.Seed)

	log.Println("Iniciando seed de dados...")
	startTime := time.Now()

	// Execute os seeds em sequência lógica (respeitando possíveis dependências)
	if err := SeedContacts(db, config.ContactsCount); err != nil {
		return err
	}

	if err := SeedUsers(db, config.UsersCount); err != nil {
		return err
	}

	if err := SeedProducts(db, config.ProductsCount); err != nil {
		return err
	}

	if err := SeedTransactions(db, config.TransactionsCount); err != nil {
		return err
	}

	if err := SeedCampaigns(db, config.CampaignsCount); err != nil {
		return err
	}

	if err := SeedRentals(db, config.RentalsCount); err != nil {
		return err
	}

	if err := SeedSales(db, config.SalesCount); err != nil {
		return err
	}

	log.Printf("Seed concluído em %v. Registros criados: %d contatos, %d usuários, %d produtos, %d transações, %d campanhas, %d aluguéis, %d vendas\n",
		time.Since(startTime),
		config.ContactsCount,
		config.UsersCount,
		config.ProductsCount,
		config.TransactionsCount,
		config.CampaignsCount,
		config.RentalsCount,
		config.SalesCount)

	return nil
}
