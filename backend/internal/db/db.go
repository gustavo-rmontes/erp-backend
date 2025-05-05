package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Driver do PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Driver do File (importante!)
	_ "github.com/lib/pq"                                      // Driver PostgreSQL para sql.Open
	"github.com/spf13/viper"

	"gorm.io/driver/postgres" // Go Orm Postgres driver
	"gorm.io/gorm"            // Go Orm
)

// OpenDB abre uma conexão com o banco de dados PostgreSQL.
func OpenDB() (*sql.DB, error) {
	// Certifica-se de que o Viper esteja lendo as variáveis do ambiente.
	viper.AutomaticEnv()

	// Obtém as variáveis de ambiente necessárias.
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbname := viper.GetString("DB_NAME")

	// Verifica se as variáveis essenciais foram definidas.
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("variáveis de ambiente do banco de dados não definidas corretamente")
	}

	// Cria a string de conexão.
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Abre a conexão com o banco.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Testa a conexão.
	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Conexão com o banco de dados estabelecida com sucesso!")
	return db, nil
}

// OpenGormDB abre uma conexão com o banco de dados usando Gorm.
func OpenGormDB() (*gorm.DB, error) {
	// Certifica-se de que o Viper esteja lendo as variáveis do ambiente.
	viper.AutomaticEnv()

	// Obtém as variáveis de ambiente necessárias.
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbname := viper.GetString("DB_NAME")

	// Verifica se as variáveis essenciais foram definidas.
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("variáveis de ambiente do banco de dados não definidas corretamente")
	}

	// Cria a string de conexão.
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Abre a conexão com o banco usando Gorm.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("[db.go]: erro ao conectar ao banco de dados com Gorm: %v", err)
	}

	log.Println("Conexão com o banco de dados via Gorm estabelecida com sucesso!")
	return db, nil
}

// RunMigrations executa as migrações do banco de dados usando variáveis de ambiente do Viper
func RunMigrations() error {
	// Garante que o Viper está lendo as variáveis de ambiente
	viper.AutomaticEnv()

	// Obtém o diretório atual de trabalho
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("erro ao obter diretório atual: %v", err)
	}

	// Ajusta o caminho das migrações para a estrutura do projeto
	migrationsPath := filepath.Join(wd, "backend", "internal", "db", "migrations")
	log.Printf("Usando diretório de migrações: %s", migrationsPath)

	// Verifica se o diretório existe
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("diretório de migrações não encontrado: %s", migrationsPath)
	}

	// Obtém os parâmetros de conexão com o banco de dados do Viper
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbname := viper.GetString("DB_NAME")

	// Verifica se as variáveis essenciais foram definidas
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return fmt.Errorf("variáveis de ambiente do banco de dados não definidas corretamente")
	}

	// Constrói a string de conexão para as migrações
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
	log.Printf("Conectando ao banco de dados: %s@%s:%s/%s", user, host, port, dbname)

	// Inicializa o migrate
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("erro ao criar instância de migrate: %v", err)
	}
	defer m.Close()

	// Verifica o estado atual antes de executar as migrações
	currentVersion, dirty, vErr := m.Version()
	if vErr == migrate.ErrNilVersion {
		log.Printf("Banco de dados não possui versão (primeira execução)")
	} else if vErr != nil {
		log.Printf("Erro ao verificar versão atual: %v", vErr)
	} else {
		log.Printf("Versão atual do banco de dados: %d, Dirty: %v", currentVersion, dirty)
	}

	// Se o banco de dados estiver em estado 'dirty', tenta forçar a versão
	if dirty {
		log.Printf("Banco de dados em estado 'dirty' na versão %d. Tentando forçar a versão...", currentVersion)
		if err := m.Force(int(currentVersion)); err != nil {
			log.Printf("Erro ao forçar versão %d: %v", currentVersion, err)
		}
	}

	// Executa as migrações
	log.Printf("Iniciando execução das migrações...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Erro ao executar migrações: %v", err)
	} else if err == migrate.ErrNoChange {
		log.Printf("Banco de dados já está na versão mais recente")
	} else {
		log.Printf("Migrações aplicadas com sucesso")
	}

	return nil
}
