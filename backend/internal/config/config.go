package config

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config é a estrutura que armazena todas as configurações do sistema.
type Config struct {
	Port             string
	Env              string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	TokenExpiresIn   time.Duration
	RefreshExpiresIn time.Duration
	// Outras configurações podem ser adicionadas aqui
}

// LoadConfig carrega as configurações a partir do arquivo .env e das variáveis de ambiente.
func LoadConfig() (*Config, error) {
	// Obtém o diretório atual onde o comando foi executado
	wd, err := os.Getwd()
	if err != nil {
		log.Println("❌ [config.go]: Erro ao obter o diretório atual.")
		return nil, err
	}

	// Caminho absoluto do .env na raiz do projeto
	dotenvPath := filepath.Join(wd, ".env")

	// Tenta carregar o .env
	if err := godotenv.Load(dotenvPath); err != nil {
		log.Printf("⚠️ [config.go]: .env não encontrado em %s, usando variáveis de ambiente.\n", dotenvPath)
	}

	// Habilita o Viper para capturar variáveis de ambiente automaticamente.
	viper.AutomaticEnv()

	// Define valores padrão para as variáveis, caso não estejam definidas.
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "erp_user")
	viper.SetDefault("DB_PASSWORD", "changeme")
	viper.SetDefault("DB_NAME", "erp_db")
	viper.SetDefault("JWT_SECRET", "changemejwtkey")
	viper.SetDefault("TOKEN_EXPIRES_IN", "15m")
	viper.SetDefault("REFRESH_EXPIRES_IN", "7d")

	// Cria a instância de configuração
	cfg := &Config{
		Port:             viper.GetString("PORT"),
		Env:              viper.GetString("ENV"),
		DBHost:           viper.GetString("DB_HOST"),
		DBPort:           viper.GetString("DB_PORT"),
		DBUser:           viper.GetString("DB_USER"),
		DBPassword:       viper.GetString("DB_PASSWORD"),
		DBName:           viper.GetString("DB_NAME"),
		JWTSecret:        viper.GetString("JWT_SECRET"),
		TokenExpiresIn:   viper.GetDuration("TOKEN_EXPIRES_IN"),
		RefreshExpiresIn: viper.GetDuration("REFRESH_EXPIRES_IN"),
	}

	return cfg, nil
}
