package service

import (
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/accounting/models"
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

	// Inicializa o logger para os testes (evita panic se o logger for chamado)
	l, err := logger.InitLogger()
	if err != nil {
		panic("Erro ao inicializar logger: " + err.Error())
	}
	logger.Logger = l

	os.Exit(m.Run())
}

// TestAddTransaction valida a criação de uma transação.
func TestAddTransaction(t *testing.T) {
	trans := models.Transaction{
		Description: "Compra",
		Amount:      50,
		Date:        "02/01/2023", // Data no formato dd/mm/yyyy, conforme definido no modelo
	}

	added, err := AddTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao adicionar transação: %v", err)
	}
	if added.ID == 0 {
		t.Errorf("Transação não adicionada corretamente: ID retornado é zero")
	}
	if added.Description != trans.Description || added.Amount != trans.Amount {
		t.Errorf("Dados da transação adicionada divergentes: esperado %v, obtido %v", trans, added)
	}
}

// TestListTransactions valida se a listagem de transações retorna a transação adicionada.
func TestListTransactions(t *testing.T) {
	// Adiciona uma transação para garantir que haja ao menos um registro
	trans := models.Transaction{
		Description: "Compra List",
		Amount:      100,
		Date:        "03/01/2023",
	}
	added, err := AddTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao adicionar transação: %v", err)
	}

	list, err := ListTransactions()
	if err != nil {
		t.Fatalf("Erro ao listar transações: %v", err)
	}

	// Procura a transação adicionada na lista
	found := false
	for _, tx := range list {
		if tx.ID == added.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Transação adicionada não encontrada na listagem")
	}
}

// TestModifyTransaction valida a atualização de uma transação.
func TestModifyTransaction(t *testing.T) {
	// Cria uma transação inicial
	trans := models.Transaction{
		Description: "Compra Teste",
		Amount:      75,
		Date:        "04/01/2023",
	}
	added, err := AddTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao adicionar transação: %v", err)
	}

	// Dados para atualização
	newData := models.Transaction{
		Description: "Compra Atualizada",
		Amount:      80,
		Date:        "05/01/2023",
	}
	updated, err := ModifyTransaction(added.ID, newData)
	if err != nil {
		t.Fatalf("Erro ao atualizar transação: %v", err)
	}
	if updated.ID != added.ID {
		t.Errorf("ID da transação atualizada diverge: esperado %d, obtido %d", added.ID, updated.ID)
	}
	if updated.Description != newData.Description || updated.Amount != newData.Amount || updated.Date != newData.Date {
		t.Errorf("Dados atualizados divergentes: esperado %v, obtido %v", newData, updated)
	}
}

// TestRemoveTransaction valida a remoção de uma transação.
func TestRemoveTransaction(t *testing.T) {
	// Cria uma transação para remoção
	trans := models.Transaction{
		Description: "Compra Remover",
		Amount:      150,
		Date:        "06/01/2023",
	}
	added, err := AddTransaction(trans)
	if err != nil {
		t.Fatalf("Erro ao adicionar transação: %v", err)
	}

	// Remove a transação
	err = RemoveTransaction(added.ID)
	if err != nil {
		t.Errorf("Erro ao remover transação: %v", err)
	}

	// Tenta remover novamente para confirmar que não existe (deve retornar erro)
	err = RemoveTransaction(added.ID)
	if err == nil {
		t.Errorf("Esperado erro ao remover transação inexistente, mas erro não ocorreu")
	}
}
