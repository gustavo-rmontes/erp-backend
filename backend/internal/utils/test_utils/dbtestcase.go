package testutils

import (
	"database/sql"
	"testing"

	"ERP-ONSMART/backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBTest é uma estrutura auxiliar para testes que usam banco de dados
type DBTest struct {
	DB        *sql.DB
	GormDB    *gorm.DB
	T         *testing.T
	CleanupFn func()
}

// NewDBTest cria uma nova instância de DBTest
func NewDBTest(t *testing.T) *DBTest {
	// Inicializa o ambiente se necessário
	if err := InitTestEnvironment(); err != nil {
		t.Fatalf("Erro ao inicializar ambiente de teste: %v", err)
	}

	// Conecta ao banco de dados
	cfg := config.LoadTestDBConfig()
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		t.Fatalf("Erro ao conectar ao banco de dados de teste: %v", err)
	}

	// Cria conexão Gorm
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		db.Close()
		t.Fatalf("Erro ao criar conexão Gorm: %v", err)
	}

	return &DBTest{
		DB:     db,
		GormDB: gormDB,
		T:      t,
		CleanupFn: func() {
			db.Close()
		},
	}
}

// Cleanup limpa os recursos utilizados pelo teste
func (dt *DBTest) Cleanup() {
	if dt.CleanupFn != nil {
		dt.CleanupFn()
	}
}
