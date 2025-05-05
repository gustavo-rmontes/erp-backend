package models

import "testing"

func TestUserModel(t *testing.T) {
	u := User{Username: "admin", Password: "senha123"}
	if u.Username == "" || u.Password == "" {
		t.Errorf("Campos obrigatórios não preenchidos")
	}
}
