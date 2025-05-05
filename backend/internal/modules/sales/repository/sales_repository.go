package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"database/sql"
	"fmt"
)

func GetAllSales() ([]models.Sale, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `
		SELECT id, product, quantity, price, customer
		FROM sales
		ORDER BY id
	`

	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var s models.Sale
		if err := rows.Scan(&s.ID, &s.Product, &s.Quantity, &s.Price, &s.Customer); err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}

	return sales, nil
}

func GetSaleByID(id int) (models.Sale, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Sale{}, err
	}
	defer conn.Close()

	query := `
		SELECT id, product, quantity, price, customer
		FROM sales
		WHERE id = $1
	`

	var sale models.Sale
	err = conn.QueryRow(query, id).Scan(
		&sale.ID,
		&sale.Product,
		&sale.Quantity,
		&sale.Price,
		&sale.Customer,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Sale{}, fmt.Errorf("venda com ID %d não encontrada", id)
		}
		return models.Sale{}, err
	}

	return sale, nil
}

func CreateSale(s models.Sale) (models.Sale, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Sale{}, err
	}
	defer conn.Close()

	query := `
		INSERT INTO sales (product, quantity, price, customer)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err = conn.QueryRow(query, s.Product, s.Quantity, s.Price, s.Customer).Scan(&s.ID)
	if err != nil {
		return models.Sale{}, err
	}

	return s, nil
}

func UpdateSale(id int, updated models.Sale) (models.Sale, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Sale{}, err
	}
	defer conn.Close()

	query := `
		UPDATE sales
		SET product = $1,
		    quantity = $2,
		    price = $3,
		    customer = $4
		WHERE id = $5
	`

	result, err := conn.Exec(query, updated.Product, updated.Quantity, updated.Price, updated.Customer, id)
	if err != nil {
		return models.Sale{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Sale{}, err
	}

	if rowsAffected == 0 {
		return models.Sale{}, sql.ErrNoRows
	}

	updated.ID = id
	return updated, nil
}

func DeleteSale(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM sales WHERE id = $1`

	result, err := conn.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venda com ID %d não encontrado", id)
	}

	return nil
}
