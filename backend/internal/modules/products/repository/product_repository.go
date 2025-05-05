package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/products/models"
	"fmt"

	"gorm.io/gorm"
)

func CreateProduct(p *models.Product) error {
	conn, err := db.OpenGormDB()
	if err != nil {
		return err
	}

	// Certifique-se de associar o modelo à tabela
	if err := conn.Model(&models.Product{}).Create(&p).Error; err != nil {
		return err
	}
	return nil
}

func GetAllProducts() ([]models.Product, error) {
	conn, err := db.OpenGormDB()
	if err != nil {
		return nil, err
	}

	var products []models.Product
	if err := conn.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func GetProductByID(id int) (*models.Product, error) {
	conn, err := db.OpenGormDB()
	if err != nil {
		return nil, err
	}

	var product models.Product
	if err := conn.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("produto com ID %d não encontrado", id)
		}
		return nil, err
	}

	return &product, nil
}

func UpdateProductByID(id int, updated models.Product) error {
	conn, err := db.OpenGormDB()
	if err != nil {
		return err
	}

	if err := conn.Model(&models.Product{}).Where("id = ?", id).Updates(updated).Error; err != nil {
		return err
	}

	// Verifica se o produto foi encontrado
	var count int64
	conn.Model(&models.Product{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return fmt.Errorf("produto com ID %d não encontrado", id)
	}
	return nil
}

func DeleteProductByID(id int) error {
	conn, err := db.OpenGormDB()
	if err != nil {
		return err
	}

	// Hard delete direto por chave primária
	result := conn.Unscoped().Delete(&models.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("produto com ID %d não encontrado", id)
	}

	return nil
}
