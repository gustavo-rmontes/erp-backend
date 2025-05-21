package testutils

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"ERP-ONSMART/backend/internal/config"
)

// SetupTestDB cria um banco de dados de teste se não existir
func SetupTestDB() error {
	cfg := config.LoadTestDBConfig()

	// Conectar ao banco 'postgres' (que sempre existe) em vez de não especificar
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}
	defer db.Close()

	// Verificar se o banco de dados de teste já existe
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("erro ao verificar existência do banco de dados: %w", err)
	}

	// Se não existir, criar o banco de dados
	if !exists {
		_, err = db.Exec("CREATE DATABASE " + cfg.DBName)
		if err != nil {
			return fmt.Errorf("erro ao criar banco de dados de teste: %w", err)
		}
		log.Printf("Banco de dados de teste '%s' criado com sucesso", cfg.DBName)
	} else {
		log.Printf("Banco de dados de teste '%s' já existe", cfg.DBName)
	}

	return nil
}

// TearDownTestDB remove o banco de dados de teste
func TearDownTestDB() error {
	cfg := config.LoadTestDBConfig()

	// Conectar ao banco 'postgres' (que sempre existe) em vez de não especificar
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}
	defer db.Close()

	// Forçar desconexão de todos os clientes
	query := fmt.Sprintf("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()", cfg.DBName)
	_, err = db.Exec(query)
	if err != nil {
		log.Printf("Aviso: erro ao desconectar clientes: %v", err)
	}

	// Excluir o banco de dados
	_, err = db.Exec("DROP DATABASE IF EXISTS " + cfg.DBName)
	if err != nil {
		return fmt.Errorf("erro ao excluir banco de dados de teste: %w", err)
	}

	log.Printf("Banco de dados de teste '%s' removido com sucesso", cfg.DBName)
	return nil
}

// CleanTestTables limpa todas as tabelas do banco de dados de teste
func CleanTestTables() error {
	cfg := config.LoadTestDBConfig()

	// Conectar ao banco de teste
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return fmt.Errorf("erro ao conectar ao banco de dados de teste: %w", err)
	}
	defer db.Close()

	// Obter todas as tabelas do esquema public
	rows, err := db.Query(`
        SELECT tablename FROM pg_tables 
        WHERE schemaname = 'public'
    `)
	if err != nil {
		return fmt.Errorf("erro ao listar tabelas: %w", err)
	}
	defer rows.Close()

	// Construir e executar comandos TRUNCATE para cada tabela
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("erro ao ler nome da tabela: %w", err)
		}
		tables = append(tables, tableName)
	}

	if len(tables) > 0 {
		// Desativa temporariamente as constraints de chave estrangeira
		if _, err := db.Exec("SET session_replication_role = 'replica';"); err != nil {
			return fmt.Errorf("erro ao desativar constraints: %w", err)
		}

		// Trunca todas as tabelas de uma vez
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;",
			strings.Join(tables, ", "))
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("erro ao limpar tabelas: %w", err)
		}

		// Reativa as constraints
		if _, err := db.Exec("SET session_replication_role = 'origin';"); err != nil {
			return fmt.Errorf("erro ao reativar constraints: %w", err)
		}

		log.Printf("Limpou %d tabelas no banco de dados de teste", len(tables))
	} else {
		log.Printf("Nenhuma tabela encontrada para limpar")
	}

	return nil
}
