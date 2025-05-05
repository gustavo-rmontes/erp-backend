package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/auth/models"
	"os"
	"testing"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env") // ← Caminho ajustado
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

func cleanupTestUser(username string) {
	conn, _ := db.OpenDB()
	defer conn.Close()
	conn.Exec("DELETE FROM users WHERE username = $1", username)
}

func TestInsertAndFindUser(t *testing.T) {
	username := "test_repo_user"
	rawPassword := "123456"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	user := models.User{Username: username, Password: string(hashed)}
	err := InsertUser(user)
	if err != nil {
		t.Fatalf("Erro ao inserir usuário no banco: %v", err)
	}
	defer cleanupTestUser(username)

	found, err := FindUserByUsername(username)
	if err != nil {
		t.Fatalf("Erro ao buscar usuário: %v", err)
	}
	if found.Username != username {
		t.Errorf("Esperado username %s, obtido %s", username, found.Username)
	}
}

func TestGetProfile(t *testing.T) {
	username := "test_profile_user"
	rawPassword := "123456"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	user := models.User{Username: username, Password: string(hashed)}
	err := InsertUser(user)
	if err != nil {
		t.Fatalf("Erro ao inserir usuário para perfil: %v", err)
	}
	defer cleanupTestUser(username)

	profile, err := GetProfile(username)
	if err != nil {
		t.Fatalf("Erro ao buscar perfil: %v", err)
	}
	if profile.Username != username {
		t.Errorf("Perfil retornado incorreto")
	}
}

func TestDeleteUserByUsername(t *testing.T) {
	username := "test_delete_user"
	rawPassword := "senha123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	user := models.User{Username: username, Password: string(hashed)}
	err := InsertUser(user)
	if err != nil {
		t.Fatalf("Erro ao inserir usuário para deleção: %v", err)
	}

	err = DeleteUserByUsername(username)
	if err != nil {
		t.Fatalf("Erro ao deletar usuário: %v", err)
	}

	_, err = FindUserByUsername(username)
	if err == nil {
		t.Errorf("Usuário ainda existe após deleção")
	}
}
