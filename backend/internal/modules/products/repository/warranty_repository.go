package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/products/models"
	"database/sql"
	"fmt"
)

// CreateWarranty insere uma nova garantia no banco.
func CreateWarranty(w models.Warranty) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`INSERT INTO warranties (product_id, duration_months, price) VALUES ($1, $2, $3)`,
		w.ProductID, w.DurationMonths, w.Price)
	return err
}

// GetWarrantyByID recupera uma garantia pelo seu ID.
func GetWarrantyByID(id int) (*models.Warranty, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var w models.Warranty
	err = conn.QueryRow(`SELECT id, product_id, duration_months, price FROM warranties WHERE id = $1`, id).
		Scan(&w.ID, &w.ProductID, &w.DurationMonths, &w.Price)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// GetWarranties retorna todas as garantias cadastradas.
func GetWarranties() ([]models.Warranty, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`SELECT id, product_id, duration_months, price FROM warranties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warranties []models.Warranty
	for rows.Next() {
		var w models.Warranty
		if err := rows.Scan(&w.ID, &w.ProductID, &w.DurationMonths, &w.Price); err != nil {
			return nil, err
		}
		warranties = append(warranties, w)
	}
	return warranties, nil
}

// UpdateWarrantyByID atualiza uma garantia com base em seu ID.
func UpdateWarrantyByID(id int, updated models.Warranty) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	res, err := conn.Exec(`UPDATE warranties SET product_id=$1, duration_months=$2, price=$3 WHERE id=$4`,
		updated.ProductID, updated.DurationMonths, updated.Price, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteWarrantyByID remove uma garantia com base em seu ID.
func DeleteWarrantyByID(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := conn.Exec(`DELETE FROM warranties WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("garantia com ID %d não encontrada", id)
	}
	return nil
}
