package service

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"ERP-ONSMART/backend/internal/modules/products/repository"
	"log"
)

func CreateProduct(p *models.Product) error {
	return repository.CreateProduct(p)
}

func ListProducts() ([]models.Product, error) {
	return repository.GetAllProducts()
}

func ListProductByID(id int) (*models.Product, error) {
	return repository.GetProductByID(id)
}

func UpdateProduct(id int, updated models.Product) error {
	return repository.UpdateProductByID(id, updated)
}

func DeleteProduct(id int) error {
	err := repository.DeleteProductByID(id)
	if err != nil {
		log.Fatalf("[prod/service]: Erro ao deletar produto com ID: %d, erro: %v", id, err)
	}
	return err
}
