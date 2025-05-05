package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/marketing/models"
	"database/sql"
	"fmt"
)

func GetAllCampaigns() ([]models.Campaign, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `
		SELECT id, title, description, budget, start_date, end_date
		FROM campaigns
		ORDER BY id
	`

	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var startDate, endDate string

		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Budget, &startDate, &endDate); err != nil {
			return nil, err
		}

		c.StartDate = startDate
		c.EndDate = endDate
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

func CreateCampaign(c models.Campaign) (models.Campaign, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Campaign{}, err
	}
	defer conn.Close()

	query := `
		INSERT INTO campaigns (title, description, budget, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err = conn.QueryRow(query, c.Title, c.Description, c.Budget, c.StartDate, c.EndDate).Scan(&c.ID)
	if err != nil {
		return models.Campaign{}, err
	}

	return c, nil
}

func UpdateCampaign(id int, updated models.Campaign) (models.Campaign, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Campaign{}, err
	}
	defer conn.Close()

	query := `
		UPDATE campaigns
		SET title = $1,
		    description = $2,
		    budget = $3,
		    start_date = $4,
		    end_date = $5
		WHERE id = $6
	`

	result, err := conn.Exec(query, updated.Title, updated.Description, updated.Budget, updated.StartDate, updated.EndDate, id)
	if err != nil {
		return models.Campaign{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Campaign{}, err
	}

	if rowsAffected == 0 {
		return models.Campaign{}, sql.ErrNoRows
	}

	updated.ID = id
	return updated, nil
}

func DeleteCampaign(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM campaigns WHERE id = $1`

	result, err := conn.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Campanha com ID %d não encontrado", id)
	}

	return nil
}
