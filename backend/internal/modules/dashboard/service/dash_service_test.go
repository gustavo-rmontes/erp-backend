package service

import "testing"

func TestListDashboardModules(t *testing.T) {
	modules := ListDashboardModules()
	if len(modules) == 0 {
		t.Error("Lista de módulos do dashboard está vazia")
	}
}
