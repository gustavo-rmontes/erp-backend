package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/auth/models"
	"fmt"
)

// FindUserByUsername busca um usuário pelo username e retorna senha também.
func FindUserByUsername(username string) (models.User, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.User{}, err
	}
	defer conn.Close()

	var user models.User
	err = conn.QueryRow(`
		SELECT username, password, email, nome, telefone, cargo 
		FROM users WHERE username = $1`, username).
		Scan(&user.Username, &user.Password, &user.Email, &user.Nome, &user.Telefone, &user.Cargo)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// InsertUser insere um novo usuário no banco.
func InsertUser(user models.User) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`
		INSERT INTO users (username, password, email, nome, telefone, cargo)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.Username, user.Password, user.Email, user.Nome, user.Telefone, user.Cargo)
	return err
}

// GetProfile retorna o perfil do usuário (sem senha).
func GetProfile(username string) (models.User, error) {
	conn, err := db.OpenDB()
	if err != nil {
		return models.User{}, err
	}
	defer conn.Close()

	var user models.User
	err = conn.QueryRow(`
		SELECT username, email, nome, telefone, cargo 
		FROM users WHERE username = $1`, username).
		Scan(&user.Username, &user.Email, &user.Nome, &user.Telefone, &user.Cargo)
	return user, err
}

// DeleteUserByUsername remove um usuário do banco de dados pelo username.
func DeleteUserByUsername(username string) error {
	conn, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer conn.Close()

	var exists bool
	err = conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("usuário '%s' não encontrado", username)
	}

	_, err = conn.Exec(`DELETE FROM users WHERE username = $1`, username)
	return err
}
