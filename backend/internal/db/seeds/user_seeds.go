package seeds

import (
	"database/sql"
	"fmt"
	"log"

	"ERP-ONSMART/backend/internal/modules/auth/models"

	"github.com/brianvoe/gofakeit/v7"
)

// SeedUsers gera usuários fictícios
func SeedUsers(db *sql.DB, count int) error {
	log.Printf("[seeds:users] Iniciando geração de %d usuários...", count)

	// Verificar se a tabela users existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:users] Erro ao verificar existência da tabela 'users': %w", err)
	}

	if !exists {
		log.Printf("[seeds:users] Tabela 'users' não existe. Seed de usuários será ignorado.")
		return nil
	}

	stmt, err := db.Prepare(`
        INSERT INTO users 
        (username, password, email, nome, telefone, cargo, created_at) 
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `)
	if err != nil {
		return fmt.Errorf("[seeds:users] Erro ao preparar inserção de usuários: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:users] Inserção preparada com sucesso.")

	for i := 0; i < count; i++ {
		var user models.User
		for {
			// Gerar dados fictícios para o usuário
			user = models.User{
				Username: gofakeit.Username(),
				Password: gofakeit.Password(true, true, true, true, false, 12),
				Email:    gofakeit.Email(),
				Nome:     gofakeit.Name(),
				Telefone: gofakeit.Phone(),
				Cargo:    gofakeit.JobTitle(),
			}

			// Verificar se o nome de usuário já existe no banco
			usernameExists, err := checkUsernameExists(db, user.Username)
			if err != nil {
				return fmt.Errorf("[seeds:users] Erro ao verificar existência do username '%s': %w", user.Username, err)
			}

			if usernameExists {
				continue // Gerar outro nome de usuário
			}

			// Tentar inserir o usuário
			_, err = stmt.Exec(user.Username, user.Password, user.Email, user.Nome, user.Telefone, user.Cargo)
			if err != nil {
				return fmt.Errorf("[seeds:users] Erro ao inserir usuário #%d: %w", i+1, err)
			}

			// Se inserção for bem-sucedida, sair do loop
			break
		}
	}

	log.Printf("[seeds:users] Geração de usuários concluída com sucesso.")
	return nil
}

// checkUsernameExists verifica se o username já existe no banco de dados
func checkUsernameExists(db *sql.DB, username string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
