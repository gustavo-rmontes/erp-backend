package testutils

import (
	"log"
	"sync"

	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/db/seeds"
)

var (
	testEnvInitialized bool
	testEnvMutex       sync.Mutex
)

// InitTestEnvironment inicializa o ambiente de teste completo
func InitTestEnvironment() error {
	testEnvMutex.Lock()
	defer testEnvMutex.Unlock()

	if testEnvInitialized {
		log.Println("Ambiente de teste já inicializado")
		return nil
	}

	log.Println("Inicializando ambiente de teste...")

	// Etapa 1: Configurar banco de dados de teste
	if err := SetupTestDB(); err != nil {
		return err
	}

	// Etapa 2: Executar migrações
	if err := db.RunTestMigrations(); err != nil {
		return err
	}

	// Etapa 2.5: Limpar tabelas existentes
	if err := CleanTestTables(); err != nil {
		log.Printf("Aviso: erro ao limpar tabelas existentes: %v", err)
		// Continuamos mesmo com erro, pois pode ser primeira execução
	}

	// Etapa 3: Carregar dados de teste
	if err := seeds.SetupTestSeeds(); err != nil {
		return err
	}

	testEnvInitialized = true
	log.Println("Ambiente de teste inicializado com sucesso")
	return nil
}

// CleanupTestEnvironment limpa o ambiente de teste
func CleanupTestEnvironment() error {
	testEnvMutex.Lock()
	defer testEnvMutex.Unlock()

	if !testEnvInitialized {
		return nil
	}

	log.Println("Limpando ambiente de teste...")

	// Opção 1: Reverter migrações
	if err := db.DropTestMigrations(); err != nil {
		log.Printf("Erro ao reverter migrações: %v", err)
		// Continue com a limpeza mesmo se houver erro
	}

	// Opção 2: Remover o banco de dados de teste
	if err := TearDownTestDB(); err != nil {
		return err
	}

	testEnvInitialized = false
	log.Println("Ambiente de teste limpo com sucesso")
	return nil
}
