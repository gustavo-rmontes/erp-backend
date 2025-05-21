package repository_test

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/logger"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestEnv mantém os recursos compartilhados para testes
type TestEnv struct {
	DB     *gorm.DB
	Logger *zap.Logger
	Mock   sqlmock.Sqlmock
}

// Ambiente global de teste
var testEnv *TestEnv

// TestMain configura o ambiente para todos os testes
func TestMain(m *testing.M) {
	// Carrega variáveis de ambiente
	viper.SetConfigFile("../../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Aviso: Erro ao carregar .env: %s\n", err.Error())
	}

	// Determina se deve usar mock para testes
	useMock := viper.GetBool("TEST_USE_MOCK")

	var testDB *gorm.DB
	var err error
	var sqlDB *sql.DB
	var mock sqlmock.Sqlmock

	if useMock {
		testDB, mock, sqlDB = setupMockDatabase()
	} else {
		testDB, err = setupTestDatabase()
		if err != nil {
			fmt.Printf("Falha ao configurar banco de teste: %v\n", err)
			os.Exit(1)
		}
	}

	// Inicializa logger silencioso para testes
	testLogger := setupTestLogger()

	// Configura o ambiente de teste
	testEnv = &TestEnv{
		DB:     testDB,
		Logger: testLogger,
		Mock:   mock,
	}

	// Executa os testes
	exitCode := m.Run()

	// Limpeza ao finalizar todos os testes
	cleanupTestEnvironment()

	if sqlDB != nil {
		sqlDB.Close()
	}

	os.Exit(exitCode)
}

// setupTestDatabase configura o banco de dados para testes
func setupTestDatabase() (*gorm.DB, error) {
	// Use um banco de dados específico para testes
	originalDBName := viper.GetString("DB_NAME")
	testDBName := originalDBName + "_test"

	// Temporariamente substitui o nome do banco para usar o banco de testes
	viper.Set("DB_NAME", testDBName)
	defer viper.Set("DB_NAME", originalDBName)

	return db.OpenGormDB()
}

// setupMockDatabase configura um banco de dados mock para testes
func setupMockDatabase() (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		fmt.Printf("Falha ao criar banco de dados mock: %v\n", err)
		os.Exit(1)
	}

	// Use o dialector postgres para configurar o mock
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		fmt.Printf("Falha ao abrir GORM DB com mock: %v\n", err)
		os.Exit(1)
	}

	return gormDB, mock, sqlDB
}

// setupTestLogger configura o logger para testes
func setupTestLogger() *zap.Logger {
	return logger.WithModule("test")
}

// cleanupTestEnvironment limpa recursos após os testes
func cleanupTestEnvironment() {
	if testEnv != nil && testEnv.DB != nil {
		sqlDB, err := testEnv.DB.DB()
		if err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// RunTestInTransaction executa um teste dentro de uma transação
func RunTestInTransaction(t *testing.T, testFunc func(tx *gorm.DB)) {
	tx := testEnv.DB.Begin()
	defer tx.Rollback()
	testFunc(tx)
}

// RecordExists verifica se um registro existe na tabela
func RecordExists(t *testing.T, db *gorm.DB, tableName string, condition string, args ...interface{}) bool {
	var count int64
	result := db.Table(tableName).Where(condition, args...).Count(&count)
	assert.NoError(t, result.Error)
	return count > 0
}

// TruncateTable limpa uma tabela específica
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	result := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName))
	assert.NoError(t, result.Error)
}
