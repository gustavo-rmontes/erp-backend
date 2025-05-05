package service

import (
	"ERP-ONSMART/backend/internal/modules/auth/models"
	"ERP-ONSMART/backend/internal/modules/auth/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Authenticate verifica as credenciais do usuário.
func Authenticate(username, password string) (models.User, error) {
	user, err := repository.FindUserByUsername(username)
	if err != nil {
		return models.User{}, errors.New("usuário ou senha inválidos")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, errors.New("usuário ou senha inválidos")
	}
	return user, nil
}

// Register cria um novo usuário com senha criptografada.
func Register(user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Define cargo padrão se não vier do front
	if user.Cargo == "" {
		user.Cargo = "Colaborador"
	}

	return repository.InsertUser(user)
}

// GetUserProfile retorna o perfil do usuário pelo username.
func GetUserProfile(username string) (models.User, error) {
	return repository.GetProfile(username)
}

// DeleteUser remove um usuário pelo username.
func DeleteUser(username string) error {
	return repository.DeleteUserByUsername(username)
}
