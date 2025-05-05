package service

import (
	"ERP-ONSMART/backend/internal/modules/accounting/models"
	"ERP-ONSMART/backend/internal/modules/accounting/repository"
)

// ListTransactions retorna todas as transações ou um erro, caso ocorra.
func ListTransactions() ([]models.Transaction, error) {
	return repository.GetAllTransactions()
}

// AddTransaction adiciona uma nova transação e retorna a transação criada ou um erro.
func AddTransaction(t models.Transaction) (models.Transaction, error) {
	return repository.CreateTransaction(t)
}

// ModifyTransaction atualiza uma transação existente e retorna a transação atualizada ou um erro.
func ModifyTransaction(id int, t models.Transaction) (models.Transaction, error) {
	return repository.UpdateTransaction(id, t)
}

// RemoveTransaction remove uma transação e retorna um erro caso a remoção não ocorra.
func RemoveTransaction(id int) error {
	return repository.DeleteTransaction(id)
}
