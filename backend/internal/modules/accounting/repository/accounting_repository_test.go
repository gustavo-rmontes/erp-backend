package repository

import (
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/accounting/models"
	"fmt"
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

	// Inicializa o logger para os testes
	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

func TestCreateTransaction(t *testing.T) {
	trans := models.Transaction{
		Description: "Teste Create",
		Amount:      10,
		Date:        "2023-01-01", // Formato ISO: yyyy-mm-dd
	}

	created, err := CreateTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao criar transação: %v", err)
	}
	if created.ID == 0 {
		t.Errorf("ID da transação não foi atribuído corretamente")
	}
}

func TestGetAllTransactions(t *testing.T) {
	// Cria uma transação para garantir que haja pelo menos um registro na listagem
	trans := models.Transaction{
		Description: "Teste Get",
		Amount:      10,
		Date:        "2023-01-01", // Formato ISO: yyyy-mm-dd
	}
	created, err := CreateTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao criar transação: %v", err)
	}

	transactions, err := GetAllTransactions()
	if err != nil {
		t.Fatalf("Erro ao obter transações: %v", err)
	}

	found := false
	for _, tx := range transactions {
		if tx.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Transação criada não foi encontrada na listagem")
	}
}

// TestUpdateTransaction cria uma transação e em seguida atualiza seus dados.
func TestUpdateTransaction(t *testing.T) {
	// Cria uma transação
	trans := models.Transaction{
		Description: "Para Update",
		Amount:      20,
		Date:        "2023-01-01",
	}
	created, err := CreateTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao criar transação: %v", err)
	}

	// Dados para atualização
	newData := models.Transaction{
		Description: "Atualizado",
		Amount:      25,
		Date:        "2023-01-02",
	}
	updated, err := UpdateTransaction(created.ID, newData)
	if err != nil {
		t.Fatalf("Erro ao atualizar transação: %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("ID da transação atualizada diverge: esperado %d, obtido %d", created.ID, updated.ID)
	}
	if updated.Description != newData.Description ||
		updated.Amount != newData.Amount ||
		updated.Date != newData.Date {
		t.Errorf("Dados da transação atualizada divergentes: esperado %v, obtido %v", newData, updated)
	}
}

// TestDeleteTransaction cria uma transação e a remove, verificando se a remoção ocorreu conforme esperado.
func TestDeleteTransaction(t *testing.T) {
	// Cria uma transação para remoção
	trans := models.Transaction{
		Description: "Para Deletar",
		Amount:      30,
		Date:        "2023-01-01",
	}
	created, err := CreateTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao criar transação: %v", err)
	}

	// Remove a transação
	err = DeleteTransaction(created.ID)
	if err != nil {
		t.Errorf("Erro ao remover transação: %v", err)
	}

	// Tenta remover novamente: deve retornar erro indicando que a transação não existe
	err = DeleteTransaction(created.ID)
	if err == nil {
		t.Errorf("Esperava erro ao deletar transação inexistente, mas não houve erro")
	} else {
		expected := fmt.Sprintf("Transação com ID %d não encontrado", created.ID)
		if err.Error() != expected {
			t.Errorf("Erro inesperado ao deletar transação inexistente. Esperado: '%s', obtido: '%v'", expected, err)
		}
	}
}
