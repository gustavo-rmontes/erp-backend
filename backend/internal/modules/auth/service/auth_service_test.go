package service

import (
	"ERP-ONSMART/backend/internal/modules/auth/models"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

func TestRegisterAndAuthenticate(t *testing.T) {
	username := "testuser_service"
	password := "senha123"

	// Cria o usuário
	err := Register(models.User{Username: username, Password: password})
	if err != nil {
		t.Fatalf("Erro ao registrar: %v", err)
	}

	// Autentica
	_, err = Authenticate(username, password)
	if err != nil {
		t.Errorf("Falha na autenticação: %v", err)
	}

	// Limpa o usuário após o teste
	_ = DeleteUser(username)
}

func TestDeleteUser(t *testing.T) {
	username := "delete_service_user"
	password := "123456"

	// Registra o usuário para deletar depois
	err := Register(models.User{Username: username, Password: password})
	if err != nil {
		t.Fatalf("Erro ao registrar usuário: %v", err)
	}

	// Deleta o usuário
	err = DeleteUser(username)
	if err != nil {
		t.Fatalf("Erro ao deletar usuário: %v", err)
	}

	// Tenta autenticar e espera erro
	_, err = Authenticate(username, password)
	if err == nil {
		t.Errorf("Usuário ainda autenticável após exclusão")
	}
}
