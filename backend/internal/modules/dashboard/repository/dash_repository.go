package repository

import "ERP-ONSMART/backend/internal/modules/dashboard/models"

func GetAvailableModules() []models.DashboardModule {
	return []models.DashboardModule{
		{"marketing", "Marketing", "icon-marketing", "/marketing"},
		{"contacts", "Contatos", "icon-contacts", "/contacts"},
		{"sales", "Vendas", "icon-sales", "/sales"},
		{"dropshipping", "Dropshipping", "icon-drop", "/dropshipping"},
		{"products", "Produtos", "icon-products", "/products"},
		{"rental", "Locação", "icon-rental", "/rental"},
		{"accounting", "Financeiro", "icon-accounting", "/accounting"},
	}
}
