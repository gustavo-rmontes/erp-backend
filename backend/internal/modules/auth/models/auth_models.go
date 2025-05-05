package models

// Usado apenas para login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nome     string `json:"nome" binding:"required"`
	Telefone string `json:"telefone"` // opcional
	Cargo    string `json:"cargo"`    // default controlado no backend/admin
}
