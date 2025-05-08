package tests

import (
	"errors"
	"testing"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestCreate testa a criação de uma cotação
func TestCreate(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	// Data para o teste
	createDTO := &dto.QuotationCreate{
		ContactID:  1,
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		Notes:      "Notas de teste",
		Terms:      "Termos de teste",
		Items: []dto.QuotationItemCreate{
			convertToRealQuotationItemCreate(testQuotationItemCreate{
				ProductID:   101,
				ProductName: "Produto de Teste",
				Description: "Descrição do produto",
				Quantity:    10,
				UnitPrice:   100.00,
				Discount:    5.00,
				Tax:         15.00,
			}),
		},
	}

	// Expectativa: serviço irá chamar CreateQuotation do repositório
	mockQuotationRepo.EXPECT().
		CreateQuotation(gomock.Any()).
		DoAndReturn(func(quotation *models.Quotation) error {
			// Verificar se os dados foram transferidos corretamente para o modelo
			assert.Equal(t, createDTO.ContactID, quotation.ContactID)
			assert.Equal(t, createDTO.Notes, quotation.Notes)
			assert.Equal(t, createDTO.Terms, quotation.Terms)
			assert.Equal(t, models.QuotationStatusDraft, quotation.Status)
			assert.Equal(t, len(createDTO.Items), len(quotation.Items))

			// Simular a atribuição de ID
			quotation.ID = 1

			return nil
		})

	// Act
	result, err := service.Create(ctx, createDTO)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, createDTO.ContactID, result.ContactID)
	assert.Equal(t, createDTO.Notes, result.Notes)
	assert.Equal(t, models.QuotationStatusDraft, result.Status)
	assert.Equal(t, len(createDTO.Items), len(result.Items))
}

// TestCreate_Error testa o caso de erro na criação
func TestCreate_Error(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	createDTO := &dto.QuotationCreate{
		ContactID:  1,
		ExpiryDate: time.Now().AddDate(0, 1, 0),
		Items:      []dto.QuotationItemCreate{},
	}

	// Simular um erro no repositório
	expectedError := errors.New("erro ao criar cotação")
	mockQuotationRepo.EXPECT().
		CreateQuotation(gomock.Any()).
		Return(expectedError)

	// Act
	result, err := service.Create(ctx, createDTO)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "falha ao criar cotação")
}

// TestGetByID testa a obtenção de uma cotação por ID
func TestGetByID(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	mockQuotation := createTestQuotation(quotationID)

	// Expectativa: o repositório retornará a cotação
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(mockQuotation, nil)

	// Act
	result, err := service.GetByID(ctx, quotationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockQuotation.ID, result.ID)
	assert.Equal(t, mockQuotation.QuotationNo, result.QuotationNo)
	assert.Equal(t, mockQuotation.Status, result.Status)
	assert.Equal(t, mockQuotation.GrandTotal, result.GrandTotal)
	assert.Equal(t, len(mockQuotation.Items), len(result.Items))
}

// TestGetByID_NotFound testa o caso de não encontrar a cotação
func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 999

	// Expectativa: o repositório retornará erro
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(nil, errors.New("quotation not found"))

	// Act
	result, err := service.GetByID(ctx, quotationID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "falha ao buscar cotação")
}

// TestUpdate testa a atualização de uma cotação
func TestUpdate(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	existingQuotation := createTestQuotation(quotationID)

	updateDTO := &dto.QuotationUpdate{
		ExpiryDate: time.Now().AddDate(0, 2, 0),
		Notes:      "Notas atualizadas",
		Terms:      "Termos atualizados",
		Items: []dto.QuotationItemCreate{
			{
				BaseItem: dto.BaseItem{ // Adicionar esta linha
					ProductID:   101,
					ProductName: "Produto Atualizado",
					Description: "Descrição atualizada",
					Quantity:    20,
					UnitPrice:   110.00,
					Discount:    10.00,
					Tax:         20.00,
				}, // Fechar o bloco BaseItem
			},
		},
	}

	// Expectativa: o repositório buscará e depois atualizará a cotação
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(existingQuotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, quotation *models.Quotation) error {
			// Verificar se os dados foram atualizados corretamente
			assert.Equal(t, updateDTO.Notes, quotation.Notes)
			assert.Equal(t, updateDTO.Terms, quotation.Terms)
			assert.Equal(t, len(updateDTO.Items), len(quotation.Items))
			return nil
		})

	// Expectativa: o repositório buscará novamente a cotação após atualização
	updatedQuotation := *existingQuotation
	updatedQuotation.Notes = updateDTO.Notes
	updatedQuotation.Terms = updateDTO.Terms
	updatedQuotation.ExpiryDate = updateDTO.ExpiryDate

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(&updatedQuotation, nil)

	// Act
	result, err := service.Update(ctx, quotationID, updateDTO)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, quotationID, result.ID)
	assert.Equal(t, updateDTO.Notes, result.Notes)
	assert.Equal(t, updateDTO.Terms, result.Terms)
}

// TestDelete testa a exclusão de uma cotação
func TestDelete(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1

	// Expectativa: o repositório excluirá a cotação
	mockQuotationRepo.EXPECT().
		DeleteQuotation(quotationID).
		Return(nil)

	// Act
	err := service.Delete(ctx, quotationID)

	// Assert
	assert.NoError(t, err)
}

// TestDelete_Error testa o caso de erro na exclusão
func TestDelete_Error(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	expectedError := errors.New("erro ao excluir cotação")

	// Expectativa: o repositório retornará erro
	mockQuotationRepo.EXPECT().
		DeleteQuotation(quotationID).
		Return(expectedError)

	// Act
	err := service.Delete(ctx, quotationID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao excluir cotação")
}
