package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"ERP-ONSMART/backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunTestMigrations executa as migrações no banco de dados de teste
func RunTestMigrations() error {
	// Obtém a configuração do banco de dados de teste
	cfg := config.LoadTestDBConfig()

	// Encontra a raiz do projeto
	rootDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("erro ao encontrar raiz do projeto: %v", err)
	}

	// Constrói o caminho correto para as migrações
	migrationsPath := filepath.Join(rootDir, "backend", "internal", "db", "migrations")
	log.Printf("Usando diretório de migrações: %s", migrationsPath)

	// Verifica se o diretório existe
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("diretório de migrações não encontrado: %s", migrationsPath)
	}

	// Constrói a string de conexão para as migrações
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// Inicializa o migrate
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("erro ao criar instância de migrate: %v", err)
	}
	defer m.Close()

	// Executa as migrações
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("erro ao executar migrações: %v", err)
	}

	log.Printf("Migrações aplicadas com sucesso no banco de dados de teste")
	return nil
}

// DropTestMigrations reverte todas as migrações no banco de dados de teste
func DropTestMigrations() error {
	// Obtém a configuração do banco de dados de teste
	cfg := config.LoadTestDBConfig()

	// Encontra a raiz do projeto
	rootDir, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("erro ao encontrar raiz do projeto: %v", err)
	}

	// Constrói o caminho correto para as migrações
	migrationsPath := filepath.Join(rootDir, "backend", "internal", "db", "migrations")
	log.Printf("Usando diretório de migrações para reversão: %s", migrationsPath)

	// Constrói a string de conexão para as migrações
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// Inicializa o migrate
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("erro ao criar instância de migrate: %v", err)
	}
	defer m.Close()

	// Reverte todas as migrações
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("erro ao reverter migrações: %v", err)
	}

	log.Printf("Todas as migrações foram revertidas no banco de dados de teste")
	return nil
}

// findProjectRoot localiza a raiz do projeto procurando pelo arquivo go.mod
func findProjectRoot() (string, error) {
	// Começa pelo diretório atual
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("erro ao obter diretório atual: %v", err)
	}

	// Percorre os diretórios até encontrar o go.mod, que marca a raiz do projeto
	for {
		// Verifica se o arquivo go.mod existe neste diretório
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move para o diretório pai
		parent := filepath.Dir(dir)

		// Se chegamos à raiz do sistema de arquivos, saímos do loop
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("não foi possível encontrar a raiz do projeto (arquivo go.mod não encontrado)")
}
