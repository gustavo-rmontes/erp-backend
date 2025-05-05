package service

import (
	"ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/contact/repository"
)

func CreateContact(contact models.Contact) error {
	return repository.InsertContact(contact)
}

func ListContacts() ([]models.Contact, error) {
	return repository.GetAllContacts()
}

func RemoveContact(id int) error {
	return repository.DeleteContactByID(id)
}

func UpdateContact(id int, contact models.Contact) error {
	return repository.UpdateContactByID(id, contact)
}

func GetContact(id int) (*models.Contact, error) {
	return repository.GetContactByID(id)
}
