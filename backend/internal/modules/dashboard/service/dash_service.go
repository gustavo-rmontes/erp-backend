package service

import (
	"ERP-ONSMART/backend/internal/modules/dashboard/models"
	"ERP-ONSMART/backend/internal/modules/dashboard/repository"
)

func ListDashboardModules() []models.DashboardModule {
	return repository.GetAvailableModules()
}
