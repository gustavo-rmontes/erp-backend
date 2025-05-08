package tests

import (
	"testing"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestAddItem testa a adição de um item à cotação
func TestAddItem(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)

	newItem := convertToRealQuotationItemCreate(testQuotationItemCreate{
		ProductID:   102,
		ProductName: "Novo Produto",
		Description: "Descrição do novo produto",
		Quantity:    5,
		UnitPrice:   200.00,
		Discount:    10.00,
		Tax:         20.00,
	})

	// Expectativa: buscar cotação, atualizar e buscar novamente
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, updatedQuotation *models.Quotation) error {
			// Verificar se o item foi adicionado
			// assert.Equal(t, len(quotation.Items)+1, len(updatedQuotation.Items))
			assert.Equal(t, 2, len(updatedQuotation.Items))
			assert.Equal(t, newItem.ProductID, updatedQuotation.Items[len(updatedQuotation.Items)-1].ProductID)
			return nil
		})

	// Criar cotação atualizada para retorno
	updatedQuotation := *quotation
	updatedQuotation.Items = append(updatedQuotation.Items, models.QuotationItem{
		ID:          2,
		QuotationID: quotationID,
		ProductID:   newItem.ProductID,
		ProductName: newItem.ProductName,
		Description: newItem.Description,
		Quantity:    newItem.Quantity,
		UnitPrice:   newItem.UnitPrice,
		Discount:    newItem.Discount,
		Tax:         newItem.Tax,
		Total:       calculateItemTotal(newItem.Quantity, newItem.UnitPrice, newItem.Discount, newItem.Tax),
	})

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(&updatedQuotation, nil)

	// Act
	result, err := service.AddItem(ctx, quotationID, &newItem)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(updatedQuotation.Items), len(result.Items))
}

// TestUpdateItem testa a atualização de um item na cotação
func TestUpdateItem(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	itemID := 1
	quotation := createTestQuotation(quotationID)

	updateItem := &dto.QuotationItemCreate{
		BaseItem: dto.BaseItem{
			ProductID:   101,
			ProductName: "Produto Atualizado",
			Description: "Descrição atualizada",
			Quantity:    15,
			UnitPrice:   120.00,
			Discount:    8.00,
			Tax:         18.00,
		},
	}

	// Expectativa: buscar cotação, atualizar e buscar novamente
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, updatedQuotation *models.Quotation) error {
			// Verificar se o item foi atualizado
			if len(updatedQuotation.Items) > 0 {
				assert.Equal(t, updateItem.ProductName, updatedQuotation.Items[0].ProductName)
				assert.Equal(t, updateItem.Quantity, updatedQuotation.Items[0].Quantity)
				assert.Equal(t, updateItem.UnitPrice, updatedQuotation.Items[0].UnitPrice)
			}
			return nil
		})

	// Criar cotação atualizada para retorno
	updatedQuotation := *quotation
	updatedQuotation.Items[0].ProductName = updateItem.ProductName
	updatedQuotation.Items[0].Description = updateItem.Description
	updatedQuotation.Items[0].Quantity = updateItem.Quantity
	updatedQuotation.Items[0].UnitPrice = updateItem.UnitPrice
	updatedQuotation.Items[0].Discount = updateItem.Discount
	updatedQuotation.Items[0].Tax = updateItem.Tax
	updatedQuotation.Items[0].Total = calculateItemTotal(
		updateItem.Quantity,
		updateItem.UnitPrice,
		updateItem.Discount,
		updateItem.Tax,
	)

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(&updatedQuotation, nil)

	// Act
	result, err := service.UpdateItem(ctx, quotationID, itemID, updateItem)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updateItem.ProductName, result.Items[0].ProductName)
	assert.Equal(t, updateItem.Quantity, result.Items[0].Quantity)
}

// TestUpdateItem_NotFound testa a atualização de um item inexistente
func TestUpdateItem_NotFound(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	itemID := 999 // Item inexistente
	quotation := createTestQuotation(quotationID)

	updateItem := &dto.QuotationItemCreate{
		BaseItem: dto.BaseItem{
			ProductID:   101,
			ProductName: "Produto Atualizado",
			Description: "Descrição atualizada",
			Quantity:    15,
			UnitPrice:   120.00,
		},
	}

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	// Act
	result, err := service.UpdateItem(ctx, quotationID, itemID, updateItem)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "item não encontrado")
}

// TestRemoveItem testa a remoção de um item da cotação
func TestRemoveItem(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	itemID := 1
	quotation := createTestQuotation(quotationID)

	// Expectativa: buscar cotação, atualizar e buscar novamente
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, updatedQuotation *models.Quotation) error {
			// Verificar se o item foi removido
			assert.Equal(t, 0, len(updatedQuotation.Items))
			return nil
		})

	// Criar cotação atualizada para retorno (sem o item)
	updatedQuotation := *quotation
	updatedQuotation.Items = []models.QuotationItem{}

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(&updatedQuotation, nil)

	// Act
	result, err := service.RemoveItem(ctx, quotationID, itemID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Items))
}
