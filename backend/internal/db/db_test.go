package db

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestOpenDB(t *testing.T) {
	// Carrega o arquivo .env (ajuste o caminho se necessário).
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Log("⚠️  .env não foi carregado. Certifique-se de que as variáveis estão definidas.")
	}

	_, err := OpenDB()
	if err != nil {
		t.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}
}
