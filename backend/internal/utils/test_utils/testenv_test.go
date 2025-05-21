package testutils

import (
	"database/sql"
	"fmt"
	"testing"

	"ERP-ONSMART/backend/internal/config"

	_ "github.com/lib/pq"
)

func TestDatabaseEnvironmentExtended(t *testing.T) {
	// Primeira inicialização
	if err := InitTestEnvironment(); err != nil {
		t.Fatalf("Erro ao inicializar ambiente de teste: %v", err)
	}

	// Verifica se uma segunda inicialização é detectada corretamente
	if err := InitTestEnvironment(); err != nil {
		t.Fatalf("Erro na segunda inicialização: %v", err)
	}

	// Verificações do banco (como você já tem)
	cfg := config.LoadTestDBConfig()
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		t.Fatalf("Erro ao conectar ao banco de teste: %v", err)
	}
	defer db.Close()

	// Verifica várias tabelas para garantir que todos os seeds funcionaram
	tables := []string{"products", "users", "contacts"}
	for _, table := range tables {
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			t.Fatalf("Erro ao verificar tabela %s: %v", table, err)
		}

		if count == 0 {
			t.Errorf("Nenhum registro encontrado na tabela %s", table)
		} else {
			t.Logf("Tabela %s: %d registros", table, count)
		}
	}

	// Limpa o ambiente manualmente (não usando defer)
	if err := CleanupTestEnvironment(); err != nil {
		t.Fatalf("Erro ao limpar ambiente de teste: %v", err)
	}

	// Tenta conectar novamente para confirmar que o banco foi removido
	_, err = sql.Open("postgres", cfg.DSN())
	if err == nil {
		// Se conectar sem erro, tenta executar uma query para verificar se o banco existe
		db2, _ := sql.Open("postgres", cfg.DSN())
		defer db2.Close()

		err = db2.QueryRow("SELECT 1").Scan(&err)
		if err == nil {
			t.Errorf("Banco de dados ainda acessível após limpeza")
		}
	}
}
