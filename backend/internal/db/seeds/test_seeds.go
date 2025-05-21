// backend/internal/db/seeds/test_seeds.go
package seeds

import (
	"database/sql"
	"log"

	"ERP-ONSMART/backend/internal/config"
)

// TestSeedConfig define uma configuração padrão para testes
var TestSeedConfig = SeedConfig{
	CustomersCount:    10,
	ProductsCount:     10,
	OrdersCount:       10,
	ContactsCount:     5,
	UsersCount:        3,
	TransactionsCount: 15,
	CampaignsCount:    2,
	RentalsCount:      5,
	SalesCount:        10,
	Seed:              42, // Valor fixo para reprodutibilidade
}

// SetupTestSeeds executa o mínimo de seeds necessários para testes
func SetupTestSeeds() error {
	// Abre conexão com o banco de dados de teste
	cfg := config.LoadTestDBConfig()

	connStr := cfg.DSN()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// Executa os seeds com a configuração de teste
	if err := ExecuteSeeds(db, TestSeedConfig); err != nil {
		return err
	}

	log.Println("Seeds de teste configurados com sucesso")
	return nil
}
