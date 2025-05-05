package service

import (
	"ERP-ONSMART/backend/internal/modules/rental/models"
	"ERP-ONSMART/backend/internal/modules/rental/repository"
)

func CreateRental(r models.Rental) error {
	return repository.InsertRental(r)
}

func ListRentals() ([]models.Rental, error) {
	return repository.GetAllRentals()
}

func UpdateRental(id int, r models.Rental) error {
	return repository.UpdateRentalByID(id, r)
}

func RemoveRental(id int) error {
	return repository.DeleteRentalByID(id)
}
