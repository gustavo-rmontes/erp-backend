package service

import (
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
)

func ListSales() ([]models.Sale, error) {
	return repository.GetAllSales()
}

func GetSale(id int) (models.Sale, error) {
	return repository.GetSaleByID(id)
}

func AddSale(s models.Sale) (models.Sale, error) {
	return repository.CreateSale(s)
}

func ModifySale(id int, s models.Sale) (models.Sale, error) {
	return repository.UpdateSale(id, s)
}

func RemoveSale(id int) error {
	return repository.DeleteSale(id)
}
