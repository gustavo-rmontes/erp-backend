package logger

import "testing"

func TestInitLogger(t *testing.T) {
	logger, err := InitLogger()
	if err != nil {
		t.Fatalf("Erro ao inicializar o logger: %v", err)
	}

	// Teste simples: registrando uma mensagem de informação
	logger.Info("Logger inicializado com sucesso!")
}
