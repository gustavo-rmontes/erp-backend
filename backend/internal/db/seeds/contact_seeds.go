package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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
        (person_type, type, name, company_name, trade_name, document, secondary_doc, 
         suframa, isento, ccm, email, phone, zip_code, street, number, complement, 
         neighborhood, city, state, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:contacts] Erro ao preparar inserção de contatos: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:contacts] Inserção preparada com sucesso.")

	// Tipos possíveis para os campos
	personTypes := []string{"pf", "pj"}
	contactTypes := []string{"cliente", "fornecedor", "lead"}
	states := []string{"SP", "RJ", "MG", "RS", "PR", "SC", "BA", "GO", "DF", "PE"}
	now := time.Now()

	for i := range count {
		// Determinar o tipo de pessoa
		personType := personTypes[gofakeit.Number(0, 1)]

		// Gerar dados fictícios para o contato
		contact := models.Contact{
			PersonType:   personType,
			Type:         contactTypes[gofakeit.Number(0, 2)],
			Name:         gofakeit.Name(),
			Email:        gofakeit.Email(),
			Phone:        gofakeit.Phone(),
			ZipCode:      gofakeit.Zip(),
			Street:       gofakeit.Street(),
			Number:       fmt.Sprintf("%d", gofakeit.Number(1, 9999)),
			Complement:   gofakeit.AppName(),
			Neighborhood: gofakeit.City(),
			City:         gofakeit.City(),
			State:        states[gofakeit.Number(0, len(states)-1)],
			CreatedAt:    now,
			UpdatedAt:    now,
			Isento:       gofakeit.Bool(),
		}

		// Ajustar campos baseados no tipo de pessoa
		if personType == "pf" {
			// CPF para pessoa física (formato 999.999.999-99)
			contact.Document = fmt.Sprintf("%s.%s.%s-%s",
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("##"))

			// RG para pessoa física (formato 99.999.999-9)
			contact.SecondaryDoc = fmt.Sprintf("%s.%s.%s-%s",
				gofakeit.Numerify("##"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("#"))

			contact.CompanyName = ""
			contact.TradeName = ""
			contact.Suframa = ""
			contact.CCM = ""
		} else {
			// CNPJ para pessoa jurídica (formato 99.999.999/0001-99)
			contact.Document = fmt.Sprintf("%s.%s/%s-%s",
				gofakeit.Numerify("##"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("####"),
				gofakeit.Numerify("##"))

			// IE para pessoa jurídica (formato 999.999.999.999)
			contact.SecondaryDoc = fmt.Sprintf("%s.%s.%s.%s",
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"),
				gofakeit.Numerify("###"))

			contact.CompanyName = gofakeit.Company()
			contact.TradeName = gofakeit.AppName()

			if gofakeit.Bool() {
				contact.Suframa = gofakeit.Numerify("#########")
			} else {
				contact.Suframa = ""
			}

			if gofakeit.Bool() {
				contact.CCM = gofakeit.Numerify("########")
			} else {
				contact.CCM = ""
			}
		}

		_, err := stmt.Exec(
			contact.PersonType,
			contact.Type,
			contact.Name,
			contact.CompanyName,
			contact.TradeName,
			contact.Document,
			contact.SecondaryDoc,
			contact.Suframa,
			contact.Isento,
			contact.CCM,
			contact.Email,
			contact.Phone,
			contact.ZipCode,
			contact.Street,
			contact.Number,
			contact.Complement,
			contact.Neighborhood,
			contact.City,
			contact.State,
			contact.CreatedAt,
			contact.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("[seeds:contacts] Erro ao inserir contato #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:contacts] Geração de contatos concluída com sucesso.")
	return nil
}
