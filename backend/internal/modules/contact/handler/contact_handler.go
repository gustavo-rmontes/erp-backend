package handler

import (
	"ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/contact/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Cria um novo contato
func CreateContactHandler(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados inválidos",
			"details": err.Error(),
		})
		return
	}

	if err := service.CreateContact(contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao criar contato",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Contato criado com sucesso"})
}

// Lista todos os contatos
func ListContactsHandler(c *gin.Context) {
	contacts, err := service.ListContacts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao listar contatos",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}

// Busca um contato pelo ID
func GetContactByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	contact, err := service.GetContact(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao buscar contato",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contact": contact})
}

// Deleta um contato pelo ID
func DeleteContactHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := service.RemoveContact(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao deletar contato",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contato deletado com sucesso"})
}

// Atualiza um contato pelo ID
func UpdateContactHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados inválidos",
			"details": err.Error(),
		})
		return
	}

	if err := service.UpdateContact(id, contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "erro ao atualizar contato",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contato atualizado com sucesso"})
}
