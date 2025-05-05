package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ListSalesHandler(c *gin.Context) {
	sales, err := service.ListSales()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar vendas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sales})
}

func GetSaleHandler(c *gin.Context) {
	// Parse the ID parameter from the URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	sale, err := service.GetSale(id)
	if err != nil {
		// Check if it's "not found" error
		if err.Error() == sql.ErrNoRows.Error() || err.Error() == "venda com ID "+strconv.Itoa(id)+" não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venda não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar venda"})
		}
		return
	}

	c.JSON(http.StatusOK, sale)
}

func CreateSaleHandler(c *gin.Context) {
	var sale models.Sale

	// Tenta fazer o bind do JSON recebido
	if err := c.ShouldBindJSON(&sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valida os campos da struct
	if err := validate.Struct(sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chama o service para salvar no banco
	created, err := service.AddSale(sale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar venda"})
		return
	}

	// Loga e retorna a resposta
	logger.Logger.Info("Venda criada com sucesso", zap.Int("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

func UpdateSaleHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var sale models.Sale
	if err := c.ShouldBindJSON(&sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := service.ModifySale(id, sale)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venda não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar venda"})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

func DeleteSaleHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := service.RemoveSale(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar venda", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Venda deletado com sucesso"})
}
