package handler

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"ERP-ONSMART/backend/internal/modules/products/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateProductHandler(c *gin.Context) {
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.CreateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar produto", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func ListProductsHandler(c *gin.Context) {
	products, err := service.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar produtos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

func GetProductByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	product, err := service.ListProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func UpdateProductHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}
	if err := service.UpdateProduct(id, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar produto", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Produto atualizado com sucesso"})
}

func DeleteProductHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	log.Printf("[prod/handler]: Tentando deletar produto com ID: %d", id)

	if err := service.DeleteProduct(id); err != nil {
		log.Printf("[prod/handler]: Erro ao deletar produto com ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
		return
	}

	log.Printf("Produto com ID %d deletado com sucesso", id)
	c.JSON(http.StatusOK, gin.H{"message": "Produto deletado com sucesso"})
}
