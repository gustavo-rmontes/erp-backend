package service

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"ERP-ONSMART/backend/internal/modules/products/repository"
)

// CreateWarranty cria uma nova garantia.
func CreateWarranty(w models.Warranty) error {
	return repository.CreateWarranty(w)
}

// GetWarrantyByID recupera uma garantia pelo seu ID.
func GetWarrantyByID(id int) (*models.Warranty, error) {
	return repository.GetWarrantyByID(id)
}

// ListWarranties retorna todas as garantias.
func ListWarranties() ([]models.Warranty, error) {
	return repository.GetWarranties()
}

// UpdateWarranty atualiza uma garantia com base em seu ID.
func UpdateWarranty(id int, updated models.Warranty) error {
	return repository.UpdateWarrantyByID(id, updated)
}

// DeleteWarranty deleta uma garantia com base em seu ID.
func DeleteWarranty(id int) error {
	return repository.DeleteWarrantyByID(id)
}
