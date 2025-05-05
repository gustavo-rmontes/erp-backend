package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/rental/models"
	"fmt"
)

func InsertRental(r models.Rental) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`INSERT INTO rentals (client_name, equipment, start_date, end_date, price, billing_type) VALUES ($1, $2, $3, $4, $5, $6)`,
		r.ClientName, r.Equipment, r.StartDate, r.EndDate, r.Price, r.BillingType)
	return err
}

func GetAllRentals() ([]models.Rental, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`SELECT id, client_name, equipment, start_date, end_date, price, billing_type FROM rentals`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rentals []models.Rental
	for rows.Next() {
		var r models.Rental
		if err := rows.Scan(&r.ID, &r.ClientName, &r.Equipment, &r.StartDate, &r.EndDate, &r.Price, &r.BillingType); err != nil {
			return nil, err
		}
		rentals = append(rentals, r)
	}
	return rentals, nil
}

func UpdateRentalByID(id int, r models.Rental) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`UPDATE rentals SET client_name=$1, equipment=$2, start_date=$3, end_date=$4, price=$5, billing_type=$6 WHERE id=$7`,
		r.ClientName, r.Equipment, r.StartDate, r.EndDate, r.Price, r.BillingType, id)
	return err
}

func DeleteRentalByID(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := conn.Exec(`DELETE FROM rentals WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("locação com ID %d não encontrado", id)
	}

	return nil
}
