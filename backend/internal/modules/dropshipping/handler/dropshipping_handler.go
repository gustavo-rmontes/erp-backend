package handler

import (
	"net/http"
	"strconv"

	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/dropshipping/models"
	"ERP-ONSMART/backend/internal/modules/dropshipping/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ListDropshippingsHandler retorna todas as transações de dropshipping.
func ListDropshippingsHandler(c *gin.Context) {
	dropshippings, err := service.ListDropshippings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dropshippings})
}

// GetDropshippingHandler retorna uma transação de dropshipping pelo ID.
func GetDropshippingHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	ds, err := service.GetDropshipping(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dropshipping não encontrado"})
		return
	}
	logger.Logger.Info("Dropshipping recuperado", zap.Int("id", ds.ID))
	c.JSON(http.StatusOK, ds)
}

// CreateDropshippingHandler cria uma nova transação de dropshipping.
func CreateDropshippingHandler(c *gin.Context) {
	var ds models.Dropshipping
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validação dos dados recebidos.
	if err := validate.Struct(ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := service.AddDropshipping(ds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Logger.Info("Dropshipping criado", zap.Int("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

// UpdateDropshippingHandler atualiza uma transação de dropshipping.
func UpdateDropshippingHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var ds models.Dropshipping
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := service.ModifyDropshipping(id, ds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Logger.Info("Dropshipping atualizado", zap.Int("id", updated.ID))
	c.JSON(http.StatusOK, updated)
}

// DeleteDropshippingHandler remove uma transação de dropshipping pelo ID.
func DeleteDropshippingHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := service.RemoveDropshipping(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dropshipping não encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dropshipping excluído com sucesso"})
}
