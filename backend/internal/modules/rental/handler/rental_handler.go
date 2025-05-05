package handler

import (
	"ERP-ONSMART/backend/internal/modules/rental/models"
	"ERP-ONSMART/backend/internal/modules/rental/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateRentalHandler(c *gin.Context) {
	var r models.Rental
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.CreateRental(r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar locação", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Locação criada com sucesso"})
}

func ListRentalsHandler(c *gin.Context) {
	rentals, err := service.ListRentals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar locações", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rentals": rentals})
}

func UpdateRentalHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var r models.Rental
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.UpdateRental(id, r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar locação", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Locação atualizada com sucesso"})
}

func DeleteRentalHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := service.RemoveRental(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar locação", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Locação deletada com sucesso"})
}
