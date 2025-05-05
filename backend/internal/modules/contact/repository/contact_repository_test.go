package repository

import (
	"ERP-ONSMART/backend/internal/modules/contact/models"
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

func TestInsertAndGetContacts(t *testing.T) {
	contact := models.Contact{
		Name:  "Contato Teste",
		Email: "teste@contato.com",
		Phone: "123456789",
		Type:  "fornecedor",
	}

	err := InsertContact(contact)
	if err != nil {
		t.Fatalf("Erro ao inserir contato: %v", err)
	}

	contacts, err := GetAllContacts()
	if err != nil {
		t.Fatalf("Erro ao buscar contatos: %v", err)
	}
	if len(contacts) == 0 {
		t.Error("Lista de contatos está vazia após inserção")
	}
}

func TestUpdateContactByID(t *testing.T) {
	// Cria um novo contato para atualização
	contact := models.Contact{
		Name:  "Contato Atualizar",
		Email: "update@contato.com",
		Phone: "000000000",
		Type:  "cliente",
	}
	err := InsertContact(contact)
	if err != nil {
		t.Fatalf("Erro ao inserir contato para atualização: %v", err)
	}

	// Pega o último inserido
	contacts, _ := GetAllContacts()
	id := contacts[len(contacts)-1].ID

	// Dados atualizados
	updated := models.Contact{
		Name:  "Contato Atualizado",
		Email: "atualizado@contato.com",
		Phone: "111111111",
		Type:  "fornecedor",
	}

	err = UpdateContactByID(id, updated)
	if err != nil {
		t.Fatalf("Erro ao atualizar contato: %v", err)
	}

	// Confirma atualização
	contacts, _ = GetAllContacts()
	found := contacts[len(contacts)-1]
	if found.Name != updated.Name || found.Email != updated.Email || found.Phone != updated.Phone || found.Type != updated.Type {
		t.Errorf("Contato não foi atualizado corretamente")
	}
}

func TestDeleteContactByID(t *testing.T) {
	contact := models.Contact{
		Name:  "Apagar",
		Email: "apagar@contato.com",
		Phone: "999999999",
		Type:  "cliente",
	}
	err := InsertContact(contact)
	if err != nil {
		t.Fatalf("Erro ao inserir contato para deleção: %v", err)
	}

	contacts, _ := GetAllContacts()
	id := contacts[len(contacts)-1].ID

	err = DeleteContactByID(id)
	if err != nil {
		t.Fatalf("Erro ao deletar contato: %v", err)
	}
}

func TestDeleteContactByID_NotFound(t *testing.T) {
	// Testa a tentativa de deletar um ID inexistente
	invalidID := 999999
	err := DeleteContactByID(invalidID)
	if err == nil {
		t.Errorf("Esperado erro ao deletar contato inexistente (ID %d), mas não houve", invalidID)
	}

	// Verifica se a mensagem de erro corresponde
	expected := "contato com ID 999999 não encontrado"
	if err.Error() != expected {
		t.Errorf("Mensagem de erro inesperada. Esperado: '%s', obtido: '%s'", expected, err.Error())
	}
}
