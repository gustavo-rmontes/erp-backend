package service

import (
	"ERP-ONSMART/backend/internal/modules/dropshipping/models"
	"ERP-ONSMART/backend/internal/modules/dropshipping/repository"
	"errors"
)

// ListDropshippings retorna todas as transações de dropshipping.
func ListDropshippings() ([]models.Dropshipping, error) {
	return repository.GetAllDropshippings()
}

// GetDropshipping retorna uma transação de dropshipping pelo ID.
func GetDropshipping(id int) (models.Dropshipping, error) {
	return repository.GetDropshippingByID(id)
}

// AddDropshipping insere uma nova transação de dropshipping.
// Após a inserção, retorna o objeto criado.
func AddDropshipping(ds models.Dropshipping) (models.Dropshipping, error) {
	// Se TotalPrice não estiver informado (ou for zero), calcula como Price * Quantity.
	if ds.TotalPrice == 0 {
		ds.TotalPrice = ds.Price * float64(ds.Quantity)
	}

	// Insere no repositório.
	if err := repository.InsertDropshipping(ds); err != nil {
		return models.Dropshipping{}, err
	}

	// Para retornar o registro inserido, listamos os dropshippings e consideramos o último como o inserido.
	list, err := repository.GetAllDropshippings()
	if err != nil {
		return models.Dropshipping{}, err
	}
	if len(list) == 0 {
		return models.Dropshipping{}, errors.New("registro de dropshipping não encontrado após inserção")
	}
	return list[len(list)-1], nil
}

// ModifyDropshipping atualiza os dados de uma transação de dropshipping pelo ID.
// Após a atualização, retorna o registro atualizado.
func ModifyDropshipping(id int, ds models.Dropshipping) (models.Dropshipping, error) {
	// Recalcula TotalPrice para refletir Price * Quantity.
	ds.TotalPrice = ds.Price * float64(ds.Quantity)
	if err := repository.UpdateDropshippingByID(id, ds); err != nil {
		return models.Dropshipping{}, err
	}
	return repository.GetDropshippingByID(id)
}

// RemoveDropshipping remove uma transação de dropshipping pelo ID.
func RemoveDropshipping(id int) error {
	return repository.DeleteDropshippingByID(id)
}
