package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/marketing/models"
	"ERP-ONSMART/backend/internal/modules/marketing/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ListCampaignsHandler(c *gin.Context) {
	camps, err := service.ListCampaigns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar campanhas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": camps})
}

func CreateCampaignHandler(c *gin.Context) {
	var camp models.Campaign
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := service.AddCampaign(camp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Logger.Info("Campanha criada", zap.Int("id", created.ID))
	c.JSON(http.StatusCreated, created)
}

func UpdateCampaignHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var camp models.Campaign
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := service.ModifyCampaign(id, camp)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campanha não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar campanha"})
		}
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteCampaignHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := service.RemoveCampaign(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar campanha", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campanha deletado com sucesso"})
}
