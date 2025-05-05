package seeds

import (
	"database/sql"
	"fmt"
	"log"

	"ERP-ONSMART/backend/internal/modules/contact/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedContacts gera contatos fictícios
func SeedContacts(db *sql.DB, count int) error {
	log.Printf("[seeds:contacts] Iniciando geração de %d contatos...", count)

	// Verificar se a tabela contacts existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'contacts')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:contacts] Erro ao verificar existência da tabela 'contacts': %w", err)
	}

	if !exists {
		log.Printf("[seeds:contacts] Tabela 'contacts' não existe. Seed de contatos será ignorado.")
		return nil
	}

	// Prepare statement com sintaxe PostgreSQL
	stmt, err := db.Prepare(`
        INSERT INTO contacts 
        (name, email, phone, type) 
        VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:contacts] Erro ao preparar inserção de contatos: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:contacts] Inserção preparada com sucesso.")

	// Tipos possíveis: cliente ou fornecedor
	contactTypes := []string{"cliente", "fornecedor"}

	for i := 0; i < count; i++ {
		// Gerar dados fictícios para o contato
		contact := models.Contact{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
			Phone: gofakeit.Phone(),
			Type:  contactTypes[gofakeit.Number(0, 1)], // Alternando entre cliente e fornecedor
		}

		_, err := stmt.Exec(
			contact.Name,
			contact.Email,
			contact.Phone,
			contact.Type,
		)

		if err != nil {
			return fmt.Errorf("[seeds:contacts] Erro ao inserir contato #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:contacts] Geração de contatos concluída com sucesso.")
	return nil
}
