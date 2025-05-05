package handler

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"ERP-ONSMART/backend/internal/modules/products/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateWarrantyHandler(c *gin.Context) {
	var w models.Warranty
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.CreateWarranty(w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar garantia", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Garantia criada com sucesso"})
}

func ListWarrantiesHandler(c *gin.Context) {
	warranties, err := service.ListWarranties()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar garantias", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"warranties": warranties})
}

func UpdateWarrantyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var w models.Warranty
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.UpdateWarranty(id, w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar garantia", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Garantia atualizada com sucesso"})
}

func DeleteWarrantyHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := service.DeleteWarranty(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar garantia", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Garantia deletada com sucesso"})
}
