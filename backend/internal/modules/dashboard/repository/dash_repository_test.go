package repository

import "testing"

func TestGetAvailableModules(t *testing.T) {
	modules := GetAvailableModules()
	if len(modules) == 0 {
		t.Error("Nenhum módulo retornado")
	}
}
