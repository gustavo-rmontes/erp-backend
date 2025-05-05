package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/dropshipping/models"
	"fmt"
)

// InsertDropshipping insere uma nova transação de dropshipping no banco de dados.
func InsertDropshipping(ds models.Dropshipping) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `
		INSERT INTO dropshipping
		    (product_id, warranty_id, cliente, price, quantity, total_price, start_date, updated_at)
		VALUES 
		    ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = conn.Exec(query, ds.ProductID, ds.WarrantyID, ds.Cliente, ds.Price, ds.Quantity, ds.TotalPrice, ds.StartDate, ds.UpdatedAt)
	return err
}

// GetAllDropshippings retorna todas as transações de dropshipping.
func GetAllDropshippings() ([]models.Dropshipping, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `
		SELECT id, product_id, warranty_id, cliente, price, quantity, total_price, start_date, updated_at
		FROM dropshipping
	`
	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Dropshipping
	for rows.Next() {
		var ds models.Dropshipping
		err = rows.Scan(&ds.ID, &ds.ProductID, &ds.WarrantyID, &ds.Cliente, &ds.Price, &ds.Quantity, &ds.TotalPrice, &ds.StartDate, &ds.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, ds)
	}
	return list, nil
}

// GetDropshippingByID retorna uma transação de dropshipping pelo ID.
func GetDropshippingByID(id int) (models.Dropshipping, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Dropshipping{}, err
	}
	defer conn.Close()

	query := `
		SELECT id, product_id, warranty_id, cliente, price, quantity, total_price, start_date, updated_at
		FROM dropshipping
		WHERE id = $1
	`
	var ds models.Dropshipping
	err = conn.QueryRow(query, id).Scan(&ds.ID, &ds.ProductID, &ds.WarrantyID, &ds.Cliente, &ds.Price, &ds.Quantity, &ds.TotalPrice, &ds.StartDate, &ds.UpdatedAt)
	if err != nil {
		return ds, err
	}
	return ds, nil
}

// DeleteDropshippingByID deleta uma transação de dropshipping pelo ID.
func DeleteDropshippingByID(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM dropshipping WHERE id = $1`
	result, err := conn.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dropshipping com ID %d não encontrado", id)
	}
	return nil
}

// UpdateDropshippingByID atualiza os dados de uma transação de dropshipping.
func UpdateDropshippingByID(id int, ds models.Dropshipping) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `
		UPDATE dropshipping
		SET product_id = $1, warranty_id = $2, cliente = $3, price = $4, quantity = $5, total_price = $6, start_date = $7, updated_at = $8
		WHERE id = $9
	`
	_, err = conn.Exec(query, ds.ProductID, ds.WarrantyID, ds.Cliente, ds.Price, ds.Quantity, ds.TotalPrice, ds.StartDate, ds.UpdatedAt, id)
	return err
}
