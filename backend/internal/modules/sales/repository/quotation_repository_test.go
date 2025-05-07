package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"fmt"
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// createMockQuotationRepo creates a repository with mocked DB
func createMockQuotationRepo(db *gorm.DB) QuotationRepository {
	return &gormQuotationRepository{
		db:  db,
		log: logger.WithModule("QuotationRepository"),
	}
}

// createTestQuotation creates a test quotation with given ID
func createTestQuotation(id int) *models.Quotation {
	return &models.Quotation{
		ID:            id,
		QuotationNo:   fmt.Sprintf("TEST-QUOT-%03d", id),
		ContactID:     1,
		Status:        models.QuotationStatusDraft,
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:      1000.00,
		TaxTotal:      150.00,
		DiscountTotal: 50.00,
		GrandTotal:    1100.00,
		Notes:         "Cotação de teste",
		Terms:         "Termos de teste",
		Items: []models.QuotationItem{
			{
				ID:          1,
				QuotationID: id,
				ProductID:   1,
				ProductName: "Produto de Teste",
				ProductCode: "PROD-001",
				Description: "Descrição do produto de teste",
				Quantity:    10,
				UnitPrice:   100.00,
				Discount:    5.00,
				Tax:         15.00,
				Total:       1100.00,
			},
		},
		Contact: &contact.Contact{
			ID:         1,
			Name:       "Test Contact",
			PersonType: "pf",
			Type:       "cliente",
			Document:   "12345678901",
			Email:      "test@example.com",
			ZipCode:    "12345-678",
		},
	}
}

// TestCreateQuotation testa a criação de uma nova cotação
func TestCreateQuotation(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Create test quotation
	quotation := createTestQuotation(0) // ID 0 because it will be assigned by the DB

	// Remover o contato para evitar que o GORM tente salvá-lo
	quotation.Contact = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect quotation insert - usando ExpectQuery para pegar o RETURNING id
	mock.ExpectQuery(`INSERT INTO "quotations"`).
		WithArgs(
			sqlmock.AnyArg(), // quotation_no
			sqlmock.AnyArg(), // contact_id
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // expiry_date
			sqlmock.AnyArg(), // subtotal
			sqlmock.AnyArg(), // tax_total
			sqlmock.AnyArg(), // discount_total
			sqlmock.AnyArg(), // grand_total
			sqlmock.AnyArg(), // notes
			sqlmock.AnyArg(), // terms
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect item insert - também Query com RETURNING id
	for range quotation.Items {
		mock.ExpectQuery(`INSERT INTO "quotation_items"`).
			WithArgs(
				sqlmock.AnyArg(), // quotation_id
				sqlmock.AnyArg(), // product_id
				sqlmock.AnyArg(), // product_name
				sqlmock.AnyArg(), // product_code
				sqlmock.AnyArg(), // description
				sqlmock.AnyArg(), // quantity
				sqlmock.AnyArg(), // unit_price
				sqlmock.AnyArg(), // discount
				sqlmock.AnyArg(), // tax
				sqlmock.AnyArg(), // total
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	// Expect query to retrieve items
	mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE quotation_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			1, 1, 1, "Produto de Teste", "PROD-001",
			"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
		))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreateQuotation(quotation); err != nil {
		t.Fatalf("Error creating quotation: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify quotation ID was set
	if quotation.ID != 1 {
		t.Errorf("Expected quotation ID to be 1, got %d", quotation.ID)
	}

	// Verify items were loaded
	if len(quotation.Items) == 0 {
		t.Fatalf("Quotation items were not loaded")
	}
}

// TestGetQuotationByID tests retrieving a quotation by ID
func TestGetQuotationByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Setup test data
	testQuotation := createTestQuotation(1)
	quotationID := testQuotation.ID

	// Setup expectations for finding quotation
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1 ORDER BY "quotations"."id" LIMIT \$2`).
		WithArgs(quotationID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
			"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
		}).AddRow(
			testQuotation.ID, testQuotation.QuotationNo, testQuotation.ContactID,
			testQuotation.Status, time.Now(), time.Now(), testQuotation.ExpiryDate,
			testQuotation.SubTotal, testQuotation.TaxTotal, testQuotation.DiscountTotal,
			testQuotation.GrandTotal, testQuotation.Notes, testQuotation.Terms,
		))

	// Setup expectations for loading items - CORRIGIDO
	mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE "quotation_items"."quotation_id" = \$1`).
		WithArgs(quotationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			testQuotation.Items[0].ID, testQuotation.Items[0].QuotationID,
			testQuotation.Items[0].ProductID, testQuotation.Items[0].ProductName,
			testQuotation.Items[0].ProductCode, testQuotation.Items[0].Description,
			testQuotation.Items[0].Quantity, testQuotation.Items[0].UnitPrice,
			testQuotation.Items[0].Discount, testQuotation.Items[0].Tax,
			testQuotation.Items[0].Total,
		))

	// Setup expectations for loading contact
	mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
		WithArgs(testQuotation.ContactID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "person_type", "type", "name", "document", "email", "zip_code",
		}).AddRow(
			testQuotation.Contact.ID,
			testQuotation.Contact.PersonType,
			testQuotation.Contact.Type,
			testQuotation.Contact.Name,
			testQuotation.Contact.Document,
			testQuotation.Contact.Email,
			testQuotation.Contact.ZipCode,
		))

	// Execute function to test
	retrievedQuotation, err := repo.GetQuotationByID(quotationID)
	if err != nil {
		t.Fatalf("Error retrieving quotation: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved quotation data
	if retrievedQuotation.ID != quotationID {
		t.Errorf("Expected quotation ID %d, got %d", quotationID, retrievedQuotation.ID)
	}

	if retrievedQuotation.QuotationNo != testQuotation.QuotationNo {
		t.Errorf("Expected quotation number %s, got %s",
			testQuotation.QuotationNo, retrievedQuotation.QuotationNo)
	}

	// Verify items were loaded
	if len(retrievedQuotation.Items) == 0 {
		t.Errorf("Items were not loaded")
	}

	// Verify contact was loaded
	if retrievedQuotation.Contact == nil {
		t.Errorf("Contact was not loaded")
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1 ORDER BY "quotations"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetQuotationByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAllQuotations tests retrieving all quotations
func TestGetAllQuotations(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Test data
	totalItems := int64(3)
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quotations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
		"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
	})

	// Add 3 test quotations to results
	for i := 1; i <= int(totalItems); i++ {
		testQuot := createTestQuotation(i)
		rows.AddRow(
			testQuot.ID, testQuot.QuotationNo, testQuot.ContactID,
			testQuot.Status, time.Now(), time.Now(), testQuot.ExpiryDate,
			testQuot.SubTotal, testQuot.TaxTotal, testQuot.DiscountTotal,
			testQuot.GrandTotal, testQuot.Notes, testQuot.Terms,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "quotations" ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items and contacts for each quotation
	for i := 1; i <= int(totalItems); i++ {
		// Items query - Match exactly as GORM generates it
		mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE "quotation_items"."quotation_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "first_name", "last_name", "email",
			}).AddRow(
				1, "Test", "Contact", "test@example.com",
			))
	}

	// Execute function to test
	result, err := repo.GetAllQuotations(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all quotations: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify pagination result data
	if result.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, result.TotalItems)
	}

	if result.CurrentPage != page {
		t.Errorf("Expected page %d, got %d", page, result.CurrentPage)
	}

	if result.PageSize != pageSize {
		t.Errorf("Expected page size %d, got %d", pageSize, result.PageSize)
	}

	expectedTotalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if result.TotalPages != expectedTotalPages {
		t.Errorf("Expected total pages %d, got %d", expectedTotalPages, result.TotalPages)
	}

	// Verify quotations were returned
	quotations, ok := result.Items.([]models.Quotation)
	if !ok {
		t.Fatalf("Could not convert items to []models.Quotation")
	}

	if len(quotations) != int(totalItems) {
		t.Errorf("Expected %d quotations, got %d", totalItems, len(quotations))
	}

	// Sub-test for custom pagination
	t.Run("Custom pagination", func(t *testing.T) {
		// Setup custom pagination params
		customPage := 2
		customPageSize := 1
		params := &pagination.PaginationParams{
			Page:     customPage,
			PageSize: customPageSize,
		}

		// Calculate expected offset
		expectedOffset := pagination.CalculateOffset(customPage, customPageSize)

		// Setup expectations for count query
		mock.ExpectQuery(`SELECT count\(\*\) FROM "quotations"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

		// Setup expectations for retrieval query with custom pagination
		rows := sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
			"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
		})

		// Add just 1 quotation to results (page 2, size 1)
		testQuot := createTestQuotation(2) // Second item for page 2
		rows.AddRow(
			testQuot.ID, testQuot.QuotationNo, testQuot.ContactID,
			testQuot.Status, time.Now(), time.Now(), testQuot.ExpiryDate,
			testQuot.SubTotal, testQuot.TaxTotal, testQuot.DiscountTotal,
			testQuot.GrandTotal, testQuot.Notes, testQuot.Terms,
		)

		mock.ExpectQuery(`SELECT \* FROM "quotations" ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(customPageSize, expectedOffset).
			WillReturnRows(rows)

		// Expect queries for loading items and contacts
		// Items query - Fix: Match the exact query pattern with table name qualifier
		mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE "quotation_items"."quotation_id" = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, 2, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "first_name", "last_name", "email",
			}).AddRow(
				1, "Test", "Contact", "test@example.com",
			))

		// Execute function to test with custom pagination
		result, err := repo.GetAllQuotations(params)
		if err != nil {
			t.Fatalf("Error getting quotations with custom pagination: %v", err)
		}

		// Verify expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %v", err)
		}

		// Verify pagination result data
		if result.CurrentPage != customPage {
			t.Errorf("Expected page %d, got %d", customPage, result.CurrentPage)
		}

		if result.PageSize != customPageSize {
			t.Errorf("Expected page size %d, got %d", customPageSize, result.PageSize)
		}

		// Verify quotations count
		quotations, ok := result.Items.([]models.Quotation)
		if !ok {
			t.Fatalf("Could not convert items to []models.Quotation")
		}

		if len(quotations) != customPageSize {
			t.Errorf("Expected %d quotations, got %d", customPageSize, len(quotations))
		}
	})

	// Sub-test for invalid pagination parameters
	t.Run("Invalid pagination", func(t *testing.T) {
		// Setup invalid pagination params
		invalidParams := &pagination.PaginationParams{
			Page:     0, // Invalid page (should be >= 1)
			PageSize: 10,
		}

		// Execute function to test with invalid pagination
		result, err := repo.GetAllQuotations(invalidParams)

		// Should return error
		if err == nil {
			t.Errorf("Expected error for invalid pagination, got nil")
		}

		// Verify result is nil
		if result != nil {
			t.Errorf("Expected nil result for invalid pagination")
		}

		// Should match ErrInvalidPagination
		if err != errors.ErrInvalidPagination {
			t.Errorf("Expected ErrInvalidPagination, got %v", err)
		}
	})
}

// TestUpdateQuotation tests updating an existing quotation
func TestUpdateQuotation(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Test data
	quotationID := 1
	initialQuotation := createTestQuotation(quotationID)

	// Create updated quotation data
	updatedQuotation := &models.Quotation{
		ID:            quotationID,
		QuotationNo:   initialQuotation.QuotationNo,
		ContactID:     initialQuotation.ContactID,
		Status:        models.QuotationStatusSent,  // Updated status
		ExpiryDate:    time.Now().AddDate(0, 2, 0), // Updated expiry date
		SubTotal:      2000.00,                     // Updated values
		TaxTotal:      300.00,
		DiscountTotal: 100.00,
		GrandTotal:    2200.00,
		Notes:         "Cotação atualizada para teste",
		Terms:         "Termos atualizados",
		Items: []models.QuotationItem{
			{
				ProductID:   1,
				ProductName: "Produto Atualizado",
				ProductCode: "PROD-001",
				Description: "Descrição do produto atualizada",
				Quantity:    20,     // Updated quantity
				UnitPrice:   110.00, // Updated price
				Discount:    10.00,
				Tax:         20.00,
				Total:       2200.00,
			},
			{
				// New item
				ProductID:   2,
				ProductName: "Produto Adicional",
				ProductCode: "PROD-002",
				Description: "Descrição do produto adicional",
				Quantity:    5,
				UnitPrice:   200.00,
				Discount:    10.00,
				Tax:         30.00,
				Total:       1100.00,
			},
		},
	}

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect query to check if quotation exists
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1 ORDER BY "quotations"."id" LIMIT \$2`).
		WithArgs(quotationID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
			"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
		}).AddRow(
			initialQuotation.ID, initialQuotation.QuotationNo, initialQuotation.ContactID,
			initialQuotation.Status, time.Now(), time.Now(), initialQuotation.ExpiryDate,
			initialQuotation.SubTotal, initialQuotation.TaxTotal, initialQuotation.DiscountTotal,
			initialQuotation.GrandTotal, initialQuotation.Notes, initialQuotation.Terms,
		))

	// Expect quotation update
	mock.ExpectExec(`UPDATE "quotations" SET "id"=\$1,"quotation_no"=\$2,"contact_id"=\$3,"status"=\$4,"updated_at"=\$5,"expiry_date"=\$6,"subtotal"=\$7,"tax_total"=\$8,"discount_total"=\$9,"grand_total"=\$10,"notes"=\$11,"terms"=\$12 WHERE "id" = \$13`).
		WithArgs(
			quotationID,
			initialQuotation.QuotationNo,
			initialQuotation.ContactID,
			models.QuotationStatusSent,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			updatedQuotation.SubTotal,
			updatedQuotation.TaxTotal,
			updatedQuotation.DiscountTotal,
			updatedQuotation.GrandTotal,
			updatedQuotation.Notes,
			updatedQuotation.Terms,
			quotationID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect deletion of existing items
	mock.ExpectExec(`DELETE FROM "quotation_items" WHERE quotation_id = \$1`).
		WithArgs(quotationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// FIX: Use ExpectQuery instead of ExpectExec and include RETURNING "id"
	mock.ExpectQuery(`INSERT INTO "quotation_items" \("quotation_id","product_id","product_name","product_code","description","quantity","unit_price","discount","tax","total"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\),\(\$11,\$12,\$13,\$14,\$15,\$16,\$17,\$18,\$19,\$20\) RETURNING "id"`).
		WithArgs(
			// First item
			quotationID,
			updatedQuotation.Items[0].ProductID,
			updatedQuotation.Items[0].ProductName,
			updatedQuotation.Items[0].ProductCode,
			updatedQuotation.Items[0].Description,
			updatedQuotation.Items[0].Quantity,
			updatedQuotation.Items[0].UnitPrice,
			updatedQuotation.Items[0].Discount,
			updatedQuotation.Items[0].Tax,
			updatedQuotation.Items[0].Total,
			// Second item
			quotationID,
			updatedQuotation.Items[1].ProductID,
			updatedQuotation.Items[1].ProductName,
			updatedQuotation.Items[1].ProductCode,
			updatedQuotation.Items[1].Description,
			updatedQuotation.Items[1].Quantity,
			updatedQuotation.Items[1].UnitPrice,
			updatedQuotation.Items[1].Discount,
			updatedQuotation.Items[1].Tax,
			updatedQuotation.Items[1].Total,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdateQuotation(quotationID, updatedQuotation)
	if err != nil {
		t.Fatalf("Error updating quotation: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1 ORDER BY "quotations"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.UpdateQuotation(invalidID, updatedQuotation)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// / TestDeleteQuotation tests deleting a quotation by ID
func TestDeleteQuotation(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Test data
	quotationID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect check for related sales orders - match exact SQL pattern
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_orders" WHERE quotation_id = \$1`).
		WithArgs(quotationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect deletion of items - match exact SQL pattern
	mock.ExpectExec(`DELETE FROM "quotation_items" WHERE quotation_id = \$1`).
		WithArgs(quotationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of quotation - FIX: include table name qualifier in WHERE clause
	mock.ExpectExec(`DELETE FROM "quotations" WHERE "quotations"."id" = \$1`).
		WithArgs(quotationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeleteQuotation(quotationID)
	if err != nil {
		t.Fatalf("Error deleting quotation: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with existing related records
	quotationID = 2
	mock.ExpectBegin()

	// Expect check for related sales orders - found 1 record
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_orders" WHERE quotation_id = \$1`).
		WithArgs(quotationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect rollback since related records exist
	mock.ExpectRollback()

	// Execute function to test
	err = repo.DeleteQuotation(quotationID)
	if err == nil {
		t.Errorf("Expected error for quotation with related records, got nil")
	}

	// Verify error is ErrRelatedRecordsExist
	if fmt.Sprintf("%v", err) != fmt.Sprintf("%v: cotação possui 1 pedidos de venda associados", errors.ErrRelatedRecordsExist) {
		t.Errorf("Expected error %v, got %v", errors.ErrRelatedRecordsExist, err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetQuotationsByContact tests retrieving quotations by contact ID with pagination
func TestGetQuotationsByContact(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Test data
	contactID := 1
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query - match exact SQL with PostgreSQL placeholders
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quotations" WHERE contact_id = \$1`).
		WithArgs(contactID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
		"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
	})

	// Add test quotations to results
	for i := 1; i <= int(totalItems); i++ {
		testQuot := createTestQuotation(i)
		testQuot.ContactID = contactID
		rows.AddRow(
			testQuot.ID, testQuot.QuotationNo, testQuot.ContactID,
			testQuot.Status, time.Now(), time.Now(), testQuot.ExpiryDate,
			testQuot.SubTotal, testQuot.TaxTotal, testQuot.DiscountTotal,
			testQuot.GrandTotal, testQuot.Notes, testQuot.Terms,
		)
	}

	// Match the actual SQL query pattern with PostgreSQL placeholders
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE contact_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(contactID, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items and contacts for each quotation
	for i := 1; i <= int(totalItems); i++ {
		// Items query - match exact SQL pattern as GORM generates
		mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE "quotation_items"."quotation_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query - match exact SQL pattern
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(contactID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "first_name", "last_name", "email",
			}).AddRow(
				contactID, "Test", "Contact", "test@example.com",
			))
	}

	// Execute function to test
	result, err := repo.GetQuotationsByContact(contactID, nil)
	if err != nil {
		t.Fatalf("Error getting quotations by contact: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify pagination result data
	if result.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, result.TotalItems)
	}

	// Verify quotations were returned with correct contact ID
	quotations, ok := result.Items.([]models.Quotation)
	if !ok {
		t.Fatalf("Could not convert items to []models.Quotation")
	}

	for i, q := range quotations {
		if q.ContactID != contactID {
			t.Errorf("Quotation %d has incorrect contact ID. Expected: %d, Got: %d",
				i, contactID, q.ContactID)
		}
	}
}

// TestGetExpiredQuotations tests retrieving expired quotations with pagination
func TestGetExpiredQuotations(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockQuotationRepo(db)

	// Test data
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize
	now := time.Now()

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "quotations" WHERE expiry_date < \$1 AND status NOT IN \(\$2, \$3\)`).
		WithArgs(sqlmock.AnyArg(), models.QuotationStatusAccepted, models.QuotationStatusRejected).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Create quotation rows to return
	rows := sqlmock.NewRows([]string{
		"id", "quotation_no", "contact_id", "status", "created_at", "updated_at",
		"expiry_date", "subtotal", "tax_total", "discount_total", "grand_total", "notes", "terms",
	})

	// Add quotations with expired dates
	for i := 1; i <= int(totalItems); i++ {
		testQuot := createTestQuotation(i)
		testQuot.Status = models.QuotationStatusDraft
		testQuot.ExpiryDate = now.AddDate(0, -1, 0) // Expired 1 month ago
		rows.AddRow(
			testQuot.ID, testQuot.QuotationNo, testQuot.ContactID,
			testQuot.Status, time.Now(), time.Now(), testQuot.ExpiryDate,
			testQuot.SubTotal, testQuot.TaxTotal, testQuot.DiscountTotal,
			testQuot.GrandTotal, testQuot.Notes, testQuot.Terms,
		)
	}

	// Match the query for retrieving expired quotations
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE expiry_date < \$1 AND status NOT IN \(\$2, \$3\) ORDER BY expiry_date ASC LIMIT \$4`).
		WithArgs(sqlmock.AnyArg(), models.QuotationStatusAccepted, models.QuotationStatusRejected, pageSize).
		WillReturnRows(rows)

	// For each quotation, expect GORM to load its associations
	// Since we're returning 2 quotations, and each has items and a contact
	// we need to set up expectations for these association queries
	for i := 1; i <= int(totalItems); i++ {
		// Expect query for items
		mock.ExpectQuery(`SELECT \* FROM "quotation_items" WHERE "quotation_items"."quotation_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Expect query for contact
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))
	}

	// Since we can't reliably predict all GORM's automated behaviors like inserts and updates,
	// we'll use MatchExpectationsInOrder(false) to allow flexibility in the order of operations
	mock.MatchExpectationsInOrder(false)

	// Execute function to test
	result, err := repo.GetExpiredQuotations(nil)
	if err != nil {
		t.Fatalf("Error getting expired quotations: %v", err)
	}

	// Verify quotations were returned with correct data
	quotations, ok := result.Items.([]models.Quotation)
	if !ok {
		t.Fatalf("Could not convert items to []models.Quotation")
	}

	// Verify the number of quotations matches
	if len(quotations) != int(totalItems) {
		t.Errorf("Expected %d quotations, got %d", totalItems, len(quotations))
	}

	// Check that returned quotations have expiry dates in the past
	for i, q := range quotations {
		if !q.ExpiryDate.Before(now) {
			t.Errorf("Quotation %d is not expired. Expiry: %v, Now: %v",
				i, q.ExpiryDate, now)
		}
	}
}
