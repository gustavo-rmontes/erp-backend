package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/contact/models"
	"database/sql"
	"fmt"
)

// Insere um novo contato no banco
func InsertContact(contact models.Contact) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`
		INSERT INTO contacts (
			person_type, type, name, company_name, trade_name, document, secondary_doc, suframa, isento, ccm,
			email, phone, zip_code, street, number, complement, neighborhood, city, state
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19
		)`,
		contact.PersonType, contact.Type, contact.Name, contact.CompanyName, contact.TradeName,
		contact.Document, contact.SecondaryDoc, contact.Suframa, contact.Isento, contact.CCM,
		contact.Email, contact.Phone, contact.ZipCode, contact.Street, contact.Number,
		contact.Complement, contact.Neighborhood, contact.City, contact.State,
	)
	return err
}

// Retorna todos os contatos
func GetAllContacts() ([]models.Contact, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
		SELECT 
			id, person_type, type, name, company_name, trade_name, document, secondary_doc, suframa, isento, ccm,
			email, phone, zip_code, street, number, complement, neighborhood, city, state,
			created_at, updated_at
		FROM contacts
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var c models.Contact
		err := rows.Scan(
			&c.ID, &c.PersonType, &c.Type, &c.Name, &c.CompanyName, &c.TradeName,
			&c.Document, &c.SecondaryDoc, &c.Suframa, &c.Isento, &c.CCM,
			&c.Email, &c.Phone, &c.ZipCode, &c.Street, &c.Number,
			&c.Complement, &c.Neighborhood, &c.City, &c.State,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

// Busca um contato pelo ID
func GetContactByID(id int) (*models.Contact, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var contact models.Contact
	err = conn.QueryRow(`
        SELECT 
            id, person_type, type, name, company_name, trade_name, document, secondary_doc, suframa, isento, ccm,
            email, phone, zip_code, street, number, complement, neighborhood, city, state,
            created_at, updated_at
        FROM contacts
        WHERE id = $1
    `, id).Scan(
		&contact.ID, &contact.PersonType, &contact.Type, &contact.Name, &contact.CompanyName, &contact.TradeName,
		&contact.Document, &contact.SecondaryDoc, &contact.Suframa, &contact.Isento, &contact.CCM,
		&contact.Email, &contact.Phone, &contact.ZipCode, &contact.Street, &contact.Number,
		&contact.Complement, &contact.Neighborhood, &contact.City, &contact.State,
		&contact.CreatedAt, &contact.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contato com ID %d não encontrado", id)
		}
		return nil, err
	}

	return &contact, nil
}

// Deleta um contato pelo ID
func DeleteContactByID(id int) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := conn.Exec("DELETE FROM contacts WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contato com ID %d não encontrado", id)
	}

	return nil
}

// Atualiza os dados de um contato pelo ID
func UpdateContactByID(id int, contact models.Contact) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`
		UPDATE contacts SET 
			person_type = $1,
			type = $2,
			name = $3,
			company_name = $4,
			trade_name = $5,
			document = $6,
			secondary_doc = $7,
			suframa = $8,
			isento = $9,
			ccm = $10,
			email = $11,
			phone = $12,
			zip_code = $13,
			street = $14,
			number = $15,
			complement = $16,
			neighborhood = $17,
			city = $18,
			state = $19,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $20
	`,
		contact.PersonType, contact.Type, contact.Name, contact.CompanyName, contact.TradeName,
		contact.Document, contact.SecondaryDoc, contact.Suframa, contact.Isento, contact.CCM,
		contact.Email, contact.Phone, contact.ZipCode, contact.Street, contact.Number,
		contact.Complement, contact.Neighborhood, contact.City, contact.State,
		id,
	)
	return err
}
