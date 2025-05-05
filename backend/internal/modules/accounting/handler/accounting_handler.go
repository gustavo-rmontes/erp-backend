package handler

import (
	"net/http"
	"strconv"

	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/accounting/models"
	"ERP-ONSMART/backend/internal/modules/accounting/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ListTransactionsHandler(c *gin.Context) {
	transactions, err := service.ListTransactions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

func CreateTransactionHandler(c *gin.Context) {
	var trans models.Transaction
	if err := c.ShouldBindJSON(&trans); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(trans); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := service.AddTransaction(trans)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Logger.Info("Transação criada", zap.Int("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

func UpdateTransactionHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var trans models.Transaction
	if err := c.ShouldBindJSON(&trans); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(trans); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := service.ModifyTransaction(id, trans)
	if err != nil {
		// Se o erro for de linha não encontrada, responde com 404, senão com 500
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transação não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteTransactionHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := service.RemoveTransaction(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar transação", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transação deletado com sucesso"})
}
