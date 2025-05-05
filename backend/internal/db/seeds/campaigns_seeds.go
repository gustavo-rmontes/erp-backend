package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"ERP-ONSMART/backend/internal/modules/marketing/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedCampaigns gera campanhas de marketing fictícias
func SeedCampaigns(db *sql.DB, count int) error {
	log.Printf("[seeds:campaigns] Iniciando geração de %d campanhas de marketing...", count)

	// Verificar tabela campaigns existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'campaigns')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:campaigns] Erro ao verificar existência da tabela 'campaigns': %w", err)
	}

	if !exists {
		log.Printf("[seeds:campaigns] Tabela 'campaigns' não existe. Seed de campanhas será ignorado.")
		return nil
	}

	stmt, err := db.Prepare(`
        INSERT INTO campaigns 
        (title, description, budget, start_date, end_date) 
        VALUES ($1, $2, $3, $4, $5)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:campaigns] Erro ao preparar inserção de campanhas: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:campaigns] Inserção preparada com sucesso.")

	// Lista de possíveis títulos para campanhas
	campaignTitles := []string{
		"Promoção de Verão",
		"Black Friday",
		"Lançamento Produto X",
		"Campanha Fim de Ano",
		"Fidelização de Clientes",
		"Expansão de Mercado",
		"Mídia Social",
		"Evento Corporativo",
	}

	// Atual até um ano no futuro
	now := time.Now()

	for i := 0; i < count; i++ {
		// Gera datas com a campanha iniciando entre hoje e 6 meses no futuro
		startDate := gofakeit.DateRange(now, now.AddDate(0, 6, 0))

		// A campanha termina entre 1 mês e 1 ano após o início
		endDate := gofakeit.DateRange(
			startDate.AddDate(0, 1, 0),
			startDate.AddDate(1, 0, 0),
		)

		// Formata no padrão ISO (YYYY-MM-DD) para inserção no PostgreSQL
		formattedStartDateForDB := startDate.Format("2006-01-02")
		formattedEndDateForDB := endDate.Format("2006-01-02")

		// Uso mais seguro do ano com conversão explícita
		year := strconv.Itoa(2025 + gofakeit.Number(0, 2))
		titleIndex := gofakeit.Number(0, len(campaignTitles)-1)
		title := campaignTitles[titleIndex] + " " + year

		// Criamos o objeto Campaign apenas para uso visual/validação
		// Não vamos definir valores nos campos que não serão usados
		campaign := models.Campaign{
			Title:       title,
			Description: gofakeit.Sentence(20),
			Budget:      gofakeit.Price(1000, 50000),
			// Removido a definição de StartDate e EndDate no modelo
		}

		_, err := stmt.Exec(
			campaign.Title,
			campaign.Description,
			campaign.Budget,
			formattedStartDateForDB,
			formattedEndDateForDB,
		)

		if err != nil {
			return fmt.Errorf("[seeds:campaigns] Erro ao inserir campanha #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:campaigns] Geração de campanhas concluída com sucesso.")
	return nil
}
