package service

import (
	"ERP-ONSMART/backend/internal/modules/contact/models"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../.env") // Caminho relativo ao service_test.go
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestCreateAndListContacts(t *testing.T) {
	c := models.Contact{
		Name:  "Serviço Teste",
		Email: "servico@teste.com",
		Phone: "40028922",
		Type:  "cliente",
	}

	err := CreateContact(c)
	if err != nil {
		t.Fatalf("Erro ao criar contato: %v", err)
	}

	list, err := ListContacts()
	if err != nil {
		t.Fatalf("Erro ao listar contatos: %v", err)
	}
	if len(list) == 0 {
		t.Errorf("Lista de contatos está vazia após inserção")
	}
}

func TestUpdateContact(t *testing.T) {
	// Cria contato inicial
	c := models.Contact{
		Name:  "Contato para Atualizar",
		Email: "original@teste.com",
		Phone: "000000000",
		Type:  "cliente",
	}
	err := CreateContact(c)
	if err != nil {
		t.Fatalf("Erro ao criar contato: %v", err)
	}

	// Pega o último contato inserido
	list, _ := ListContacts()
	id := list[len(list)-1].ID

	// Dados atualizados
	updated := models.Contact{
		Name:  "Contato Atualizado Serviço",
		Email: "novo@teste.com",
		Phone: "111111111",
		Type:  "fornecedor",
	}

	err = UpdateContact(id, updated)
	if err != nil {
		t.Fatalf("Erro ao atualizar contato: %v", err)
	}

	// Confirma alteração
	list, _ = ListContacts()
	changed := list[len(list)-1]
	if changed.Name != updated.Name || changed.Email != updated.Email || changed.Phone != updated.Phone || changed.Type != updated.Type {
		t.Errorf("Contato não foi atualizado corretamente")
	}
}
