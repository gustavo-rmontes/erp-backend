package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/accounting/models"
	"database/sql"
	"fmt"
)

// GetAllTransactions retorna todas as transações armazenadas no banco.
func GetAllTransactions() ([]models.Transaction, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := `
		SELECT id, description, amount, date
		FROM acc_transaction
		ORDER BY id
	`

	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var date string

		if err := rows.Scan(&t.ID, &t.Description, &t.Amount, &date); err != nil {
			return nil, err
		}

		// Atribui a data conforme vem do banco (normalmente já no formato yyyy-mm-dd).
		t.Date = date
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// CreateTransaction insere uma nova transação e retorna a transação criada com o ID gerado.
func CreateTransaction(t models.Transaction) (models.Transaction, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Transaction{}, err
	}
	defer conn.Close()

	query := `
		INSERT INTO acc_transaction (description, amount, date)
		VALUES ($1, $2, TO_DATE($3, 'DD/MM/YYYY'))
		RETURNING id
	`

	err = conn.QueryRow(query, t.Description, t.Amount, t.Date).Scan(&t.ID)
	if err != nil {
		return models.Transaction{}, err
	}

	return t, nil
}

// UpdateTransaction atualiza os dados de uma transação existente.
func UpdateTransaction(id int, updated models.Transaction) (models.Transaction, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.Transaction{}, err
	}
	defer conn.Close()

	query := `
		UPDATE acc_transaction
		SET description = $1,
		    amount = $2,
		    date = TO_DATE($3, 'DD/MM/YYYY')
		WHERE id = $4
	`

	result, err := conn.Exec(query, updated.Description, updated.Amount, updated.Date, id)
	if err != nil {
		return models.Transaction{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Transaction{}, err
	}
	if rowsAffected == 0 {
		return models.Transaction{}, sql.ErrNoRows
	}

	updated.ID = id
	return updated, nil
}

// DeleteTransaction remove uma transação a partir de seu ID.
func DeleteTransaction(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM acc_transaction WHERE id = $1`

	result, err := conn.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("transação com ID %d não encontrado", id)
	}

	return nil
}
