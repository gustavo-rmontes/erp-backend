package handler

import (
	"ERP-ONSMART/backend/internal/modules/dashboard/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DashboardHandler(c *gin.Context) {
	modules := service.ListDashboardModules()
	c.JSON(http.StatusOK, gin.H{"modules": modules})
}
