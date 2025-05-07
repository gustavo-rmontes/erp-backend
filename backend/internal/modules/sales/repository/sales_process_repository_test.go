package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// createMockSalesProcessRepo creates a repository with mocked DB
func createMockSalesProcessRepo(db *gorm.DB) SalesProcessRepository {
	return &gormSalesProcessRepository{
		db:  db,
		log: logger.WithModule("SalesProcessRepository"),
	}
}

// createTestSalesProcess creates a test sales process with given ID
func createTestSalesProcess(id int) *models.SalesProcess {
	return &models.SalesProcess{
		ID:         id,
		ContactID:  1,
		Status:     "draft",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		TotalValue: 2000.00,
		Profit:     500.00,
		Notes:      "Processo de vendas de teste",
		Contact: &contact.Contact{
			ID:         1,
			Name:       "Test Contact",
			PersonType: "pf",
			Type:       "cliente",
			Document:   "12345678901",
			Email:      "test@example.com",
			ZipCode:    "12345-678",
		},
		Quotation: &models.Quotation{
			ID:            1,
			QuotationNo:   "TEST-QUOT-001",
			ContactID:     1,
			Status:        models.QuotationStatusAccepted,
			ExpiryDate:    time.Now().AddDate(0, 1, 0),
			SubTotal:      1000.00,
			TaxTotal:      150.00,
			DiscountTotal: 50.00,
			GrandTotal:    1100.00,
		},
		SalesOrder: &models.SalesOrder{
			ID:     1,
			SONo:   "SO-001",
			Status: models.SOStatusConfirmed,
		},
	}
}

// TestCreateSalesProcess tests the creation of a new sales process
func TestCreateSalesProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Create test sales process
	process := createTestSalesProcess(0) // ID 0 because it will be assigned by the DB

	// Remove relationships to avoid GORM trying to save them
	process.Contact = nil
	process.Quotation = nil
	process.SalesOrder = nil
	process.PurchaseOrder = nil
	process.Deliveries = nil
	process.Invoices = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect sales process insert
	mock.ExpectQuery(`INSERT INTO "sales_processes"`).
		WithArgs(
			sqlmock.AnyArg(), // contact_id
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // total_value
			sqlmock.AnyArg(), // profit
			sqlmock.AnyArg(), // notes
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreateSalesProcess(process); err != nil {
		t.Fatalf("Error creating sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify sales process ID was set
	if process.ID != 1 {
		t.Errorf("Expected sales process ID to be 1, got %d", process.ID)
	}
}

// TestGetSalesProcessByID tests retrieving a sales process by ID
func TestGetSalesProcessByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Setup test data
	testProcess := createTestSalesProcess(1)
	processID := testProcess.ID

	// Setup expectations for finding sales process
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status", "created_at", "updated_at",
			"total_value", "profit", "notes",
		}).AddRow(
			testProcess.ID, testProcess.ContactID, testProcess.Status,
			testProcess.CreatedAt, testProcess.UpdatedAt, testProcess.TotalValue,
			testProcess.Profit, testProcess.Notes,
		))

	// Setup expectations for loading contact
	mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
		WithArgs(testProcess.ContactID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "person_type", "type", "name", "document", "email", "zip_code",
		}).AddRow(
			testProcess.Contact.ID,
			testProcess.Contact.PersonType,
			testProcess.Contact.Type,
			testProcess.Contact.Name,
			testProcess.Contact.Document,
			testProcess.Contact.Email,
			testProcess.Contact.ZipCode,
		))

	// Skip further relationship loading since we can't effectively mock it
	// We'll expect these queries but return empty results so no further loading is attempted

	// Setup expectations for loading quotation relationship (only IDs)
	mock.ExpectQuery(`SELECT "quotation_id" FROM "process_quotations" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnRows(sqlmock.NewRows([]string{"quotation_id"}))

	// Setup expectations for loading sales order relationship (only IDs)
	mock.ExpectQuery(`SELECT "sales_order_id" FROM "process_sales_orders" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnRows(sqlmock.NewRows([]string{"sales_order_id"}))

	// Setup expectations for loading purchase order relationship (only IDs)
	mock.ExpectQuery(`SELECT "purchase_order_id" FROM "process_purchase_orders" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnRows(sqlmock.NewRows([]string{"purchase_order_id"}))

	// Setup expectations for loading delivery relationship (only IDs)
	mock.ExpectQuery(`SELECT "delivery_id" FROM "process_deliveries" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnRows(sqlmock.NewRows([]string{"delivery_id"}))

	// Setup expectations for loading invoice relationship (only IDs)
	mock.ExpectQuery(`SELECT "invoice_id" FROM "process_invoices" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnRows(sqlmock.NewRows([]string{"invoice_id"}))

	// Execute function to test
	retrievedProcess, err := repo.GetSalesProcessByID(processID)
	if err != nil {
		t.Fatalf("Error retrieving sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved sales process data
	if retrievedProcess.ID != processID {
		t.Errorf("Expected sales process ID %d, got %d", processID, retrievedProcess.ID)
	}

	if retrievedProcess.Status != testProcess.Status {
		t.Errorf("Expected sales process status %s, got %s",
			testProcess.Status, retrievedProcess.Status)
	}

	// Verify contact was loaded
	if retrievedProcess.Contact == nil {
		t.Errorf("Contact was not loaded")
	}

	// Note: We're not checking for quotation, sales order, etc. since
	// we're not mocking their loading behavior in this test

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetSalesProcessByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAllSalesProcesses tests retrieving all sales processes
func TestGetAllSalesProcesses(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	totalItems := int64(3)
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_processes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "contact_id", "status", "created_at", "updated_at",
		"total_value", "profit", "notes",
	})

	// Add test processes to results
	for i := 1; i <= int(totalItems); i++ {
		testProcess := createTestSalesProcess(i)
		rows.AddRow(
			testProcess.ID, testProcess.ContactID, testProcess.Status,
			testProcess.CreatedAt, testProcess.UpdatedAt, testProcess.TotalValue,
			testProcess.Profit, testProcess.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// For each process, setup expectations for loading contact and related documents
	for i := 1; i <= int(totalItems); i++ {
		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Setup expectations for loading quotation relationship
		mock.ExpectQuery(`SELECT "quotation_id" FROM "process_quotations" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"quotation_id"}))

		// Setup expectations for loading sales order relationship
		mock.ExpectQuery(`SELECT "sales_order_id" FROM "process_sales_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"sales_order_id"}))

		// Setup expectations for loading purchase order relationship
		mock.ExpectQuery(`SELECT "purchase_order_id" FROM "process_purchase_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"purchase_order_id"}))

		// Setup expectations for loading delivery relationship
		mock.ExpectQuery(`SELECT "delivery_id" FROM "process_deliveries" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"delivery_id"}))

		// Setup expectations for loading invoice relationship
		mock.ExpectQuery(`SELECT "invoice_id" FROM "process_invoices" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"invoice_id"}))
	}

	// Execute function to test
	result, err := repo.GetAllSalesProcesses(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all sales processes: %v", err)
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

	// Verify processes were returned
	processes, ok := result.Items.([]models.SalesProcess)
	if !ok {
		t.Fatalf("Could not convert items to []models.SalesProcess")
	}

	if len(processes) != int(totalItems) {
		t.Errorf("Expected %d processes, got %d", totalItems, len(processes))
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
		mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_processes"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

		// Setup expectations for retrieval query with custom pagination
		rows := sqlmock.NewRows([]string{
			"id", "contact_id", "status", "created_at", "updated_at",
			"total_value", "profit", "notes",
		})

		// Add just 1 process to results (page 2, size 1)
		testProcess := createTestSalesProcess(2)
		rows.AddRow(
			testProcess.ID, testProcess.ContactID, testProcess.Status,
			testProcess.CreatedAt, testProcess.UpdatedAt, testProcess.TotalValue,
			testProcess.Profit, testProcess.Notes,
		)

		mock.ExpectQuery(`SELECT \* FROM "sales_processes" ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
			WithArgs(customPageSize, expectedOffset).
			WillReturnRows(rows)

		// Setup expectations for loading contact
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Setup expectations for loading related documents - FIX: Updated SQL patterns to match actual queries
		mock.ExpectQuery(`SELECT "quotation_id" FROM "process_quotations" WHERE process_id = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"quotation_id"}))

		mock.ExpectQuery(`SELECT "sales_order_id" FROM "process_sales_orders" WHERE process_id = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"sales_order_id"}))

		mock.ExpectQuery(`SELECT "purchase_order_id" FROM "process_purchase_orders" WHERE process_id = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"purchase_order_id"}))

		mock.ExpectQuery(`SELECT "delivery_id" FROM "process_deliveries" WHERE process_id = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"delivery_id"}))

		mock.ExpectQuery(`SELECT "invoice_id" FROM "process_invoices" WHERE process_id = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"invoice_id"}))

		// Execute function to test with custom pagination
		result, err := repo.GetAllSalesProcesses(params)
		if err != nil {
			t.Fatalf("Error getting sales processes with custom pagination: %v", err)
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

		// Verify processes count
		processes, ok := result.Items.([]models.SalesProcess)
		if !ok {
			t.Fatalf("Could not convert items to []models.SalesProcess")
		}

		if len(processes) != customPageSize {
			t.Errorf("Expected %d processes, got %d", customPageSize, len(processes))
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
		result, err := repo.GetAllSalesProcesses(invalidParams)

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

// TestUpdateSalesProcess tests updating an existing sales process
func TestUpdateSalesProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	initialProcess := createTestSalesProcess(processID)

	// Create updated process data
	updatedProcess := &models.SalesProcess{
		ID:         processID,
		ContactID:  initialProcess.ContactID,
		Status:     "completed", // Updated status
		TotalValue: 2500.00,     // Updated value
		Profit:     700.00,      // Updated profit
		Notes:      "Processo de vendas atualizado para teste",
	}

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status", "created_at", "updated_at",
			"total_value", "profit", "notes",
		}).AddRow(
			initialProcess.ID, initialProcess.ContactID, initialProcess.Status,
			initialProcess.CreatedAt, initialProcess.UpdatedAt, initialProcess.TotalValue,
			initialProcess.Profit, initialProcess.Notes,
		))

	// Expect process update - FIX: Match the actual SQL query pattern including ID in SET clause
	mock.ExpectExec(`UPDATE "sales_processes" SET "id"=\$1,"contact_id"=\$2,"status"=\$3,"updated_at"=\$4,"total_value"=\$5,"profit"=\$6,"notes"=\$7 WHERE "id" = \$8`).
		WithArgs(
			processID, // id
			initialProcess.ContactID,
			updatedProcess.Status,
			sqlmock.AnyArg(), // updated_at
			updatedProcess.TotalValue,
			updatedProcess.Profit,
			updatedProcess.Notes,
			processID, // id for WHERE clause
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdateSalesProcess(processID, updatedProcess)
	if err != nil {
		t.Fatalf("Error updating sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.UpdateSalesProcess(invalidID, updatedProcess)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestDeleteSalesProcess tests deleting a sales process by ID
func TestDeleteSalesProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect deletion of relationships
	mock.ExpectExec(`DELETE FROM "process_quotations" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "process_sales_orders" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "process_purchase_orders" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "process_deliveries" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "process_invoices" WHERE process_id = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of process
	mock.ExpectExec(`DELETE FROM "sales_processes" WHERE "sales_processes"."id" = \$1`).
		WithArgs(processID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeleteSalesProcess(processID)
	if err != nil {
		t.Fatalf("Error deleting sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with non-existent ID
	nonExistentID := 999
	mock.ExpectBegin()

	// Expect deletion of relationships
	mock.ExpectExec(`DELETE FROM "process_quotations" WHERE process_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM "process_sales_orders" WHERE process_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM "process_purchase_orders" WHERE process_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM "process_deliveries" WHERE process_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM "process_invoices" WHERE process_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Expect deletion of process
	mock.ExpectExec(`DELETE FROM "sales_processes" WHERE "sales_processes"."id" = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectRollback()

	err = repo.DeleteSalesProcess(nonExistentID)
	if err == nil {
		t.Errorf("Expected error for non-existent ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLinkQuotationToProcess tests linking a quotation to a sales process
func TestLinkQuotationToProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	quotationID := 1

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Skip the actual quotation repository initialization and directly mock the specific query
	// that LinkQuotationToProcess will execute to check if quotation exists

	// Since the NewQuotationRepository method makes a direct DB call, we'll need to simplify our test
	// Let's just mock the behavior we want from the GetQuotationByID method

	// Instead of relying on testing database initialization or the actual repository creation,
	// focus on mocking the specific SQL queries the function executes
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
		WithArgs(quotationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status",
		}).AddRow(
			quotationID, "TEST-QUOT-001", 1, "accepted",
		))

	// Expect query to check if link already exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_quotations" WHERE process_id = \$1 AND quotation_id = \$2`).
		WithArgs(processID, quotationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect query to create link
	mock.ExpectExec(`INSERT INTO process_quotations \(process_id, quotation_id\) VALUES \(\$1, \$2\)`).
		WithArgs(processID, quotationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute function to test
	err := repo.LinkQuotationToProcess(processID, quotationID)
	if err != nil {
		t.Fatalf("Error linking quotation to sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test when link already exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Mock quotation query directly
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
		WithArgs(quotationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status",
		}).AddRow(
			quotationID, "TEST-QUOT-001", 1, "accepted",
		))

	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_quotations" WHERE process_id = \$1 AND quotation_id = \$2`).
		WithArgs(processID, quotationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Execute function to test
	err = repo.LinkQuotationToProcess(processID, quotationID)
	if err != nil {
		t.Fatalf("Error (unexpected) when link already exists: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid process ID
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err = repo.LinkQuotationToProcess(invalidID, quotationID)
	if err == nil {
		t.Errorf("Expected error for invalid process ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLinkSalesOrderToProcess tests linking a sales order to a sales process
func TestLinkSalesOrderToProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	salesOrderID := 1

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Expect query to check if sales order exists
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(salesOrderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "contact_id", "status",
		}).AddRow(salesOrderID, "SO-001", 1, "confirmed"))

	// Expect query to check if link already exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_sales_orders" WHERE process_id = \$1 AND sales_order_id = \$2`).
		WithArgs(processID, salesOrderID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect query to create link
	mock.ExpectExec(`INSERT INTO process_sales_orders \(process_id, sales_order_id\) VALUES \(\$1, \$2\)`).
		WithArgs(processID, salesOrderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute function to test
	err := repo.LinkSalesOrderToProcess(processID, salesOrderID)
	if err != nil {
		t.Fatalf("Error linking sales order to sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLinkPurchaseOrderToProcess tests linking a purchase order to a sales process
func TestLinkPurchaseOrderToProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	purchaseOrderID := 1

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Expect query to check if purchase order exists
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1 ORDER BY "purchase_orders"."id" LIMIT \$2`).
		WithArgs(purchaseOrderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "po_no", "contact_id", "status",
		}).AddRow(purchaseOrderID, "PO-001", 1, "confirmed"))

	// Expect query to check if link already exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_purchase_orders" WHERE process_id = \$1 AND purchase_order_id = \$2`).
		WithArgs(processID, purchaseOrderID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect query to create link
	mock.ExpectExec(`INSERT INTO process_purchase_orders \(process_id, purchase_order_id\) VALUES \(\$1, \$2\)`).
		WithArgs(processID, purchaseOrderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute function to test
	err := repo.LinkPurchaseOrderToProcess(processID, purchaseOrderID)
	if err != nil {
		t.Fatalf("Error linking purchase order to sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLinkDeliveryToProcess tests linking a delivery to a sales process
func TestLinkDeliveryToProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	deliveryID := 1

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Expect query to check if delivery exists
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE "deliveries"."id" = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "delivery_no", "status",
		}).AddRow(deliveryID, "DEL-001", "pending"))

	// Expect query to check if link already exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_deliveries" WHERE process_id = \$1 AND delivery_id = \$2`).
		WithArgs(processID, deliveryID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect query to create link
	mock.ExpectExec(`INSERT INTO process_deliveries \(process_id, delivery_id\) VALUES \(\$1, \$2\)`).
		WithArgs(processID, deliveryID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute function to test
	err := repo.LinkDeliveryToProcess(processID, deliveryID)
	if err != nil {
		t.Fatalf("Error linking delivery to sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLinkInvoiceToProcess tests linking an invoice to a sales process
func TestLinkInvoiceToProcess(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	processID := 1
	invoiceID := 1

	// Expect query to check if process exists
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE "sales_processes"."id" = \$1 ORDER BY "sales_processes"."id" LIMIT \$2`).
		WithArgs(processID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contact_id", "status",
		}).AddRow(processID, 1, "draft"))

	// Expect query to check if invoice exists
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(invoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "status",
		}).AddRow(invoiceID, "INV-001", "draft"))

	// Expect query to check if link already exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "process_invoices" WHERE process_id = \$1 AND invoice_id = \$2`).
		WithArgs(processID, invoiceID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect query to create link
	mock.ExpectExec(`INSERT INTO process_invoices \(process_id, invoice_id\) VALUES \(\$1, \$2\)`).
		WithArgs(processID, invoiceID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute function to test
	err := repo.LinkInvoiceToProcess(processID, invoiceID)
	if err != nil {
		t.Fatalf("Error linking invoice to sales process: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetSalesProcessByContact tests retrieving sales processes by contact ID
func TestGetSalesProcessByContact(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	contactID := 1
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_processes" WHERE contact_id = \$1`).
		WithArgs(contactID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "contact_id", "status", "created_at", "updated_at",
		"total_value", "profit", "notes",
	})

	// Add test processes to results
	for i := 1; i <= int(totalItems); i++ {
		testProcess := createTestSalesProcess(i)
		testProcess.ContactID = contactID
		rows.AddRow(
			testProcess.ID, testProcess.ContactID, testProcess.Status,
			testProcess.CreatedAt, testProcess.UpdatedAt, testProcess.TotalValue,
			testProcess.Profit, testProcess.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE contact_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(contactID, pageSize).
		WillReturnRows(rows)

	// For each process, setup expectations for loading contact and related documents
	for i := 1; i <= int(totalItems); i++ {
		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(contactID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				contactID, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Setup expectations for loading related documents - FIX: SQL patterns to match actual queries
		mock.ExpectQuery(`SELECT "quotation_id" FROM "process_quotations" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"quotation_id"}))

		mock.ExpectQuery(`SELECT "sales_order_id" FROM "process_sales_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"sales_order_id"}))

		mock.ExpectQuery(`SELECT "purchase_order_id" FROM "process_purchase_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"purchase_order_id"}))

		mock.ExpectQuery(`SELECT "delivery_id" FROM "process_deliveries" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"delivery_id"}))

		mock.ExpectQuery(`SELECT "invoice_id" FROM "process_invoices" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"invoice_id"}))
	}

	// Execute function to test
	result, err := repo.GetSalesProcessByContact(contactID, nil)
	if err != nil {
		t.Fatalf("Error getting sales processes by contact: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify processes were returned with correct contact ID
	processes, ok := result.Items.([]models.SalesProcess)
	if !ok {
		t.Fatalf("Could not convert items to []models.SalesProcess")
	}

	for i, p := range processes {
		if p.ContactID != contactID {
			t.Errorf("Process %d has incorrect contact ID. Expected: %d, Got: %d",
				i, contactID, p.ContactID)
		}
	}
}

// TestGetSalesProcessByStatus tests retrieving sales processes by status
func TestGetSalesProcessByStatus(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesProcessRepo(db)

	// Test data
	status := "draft"
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_processes" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "contact_id", "status", "created_at", "updated_at",
		"total_value", "profit", "notes",
	})

	// Add test processes to results
	for i := 1; i <= int(totalItems); i++ {
		testProcess := createTestSalesProcess(i)
		testProcess.Status = status
		rows.AddRow(
			testProcess.ID, testProcess.ContactID, testProcess.Status,
			testProcess.CreatedAt, testProcess.UpdatedAt, testProcess.TotalValue,
			testProcess.Profit, testProcess.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "sales_processes" WHERE status = \$1 ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// For each process, setup expectations for loading contact and related documents
	for i := 1; i <= int(totalItems); i++ {
		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Setup expectations for loading related documents
		// FIX: Update SQL patterns to match the actual queries with proper quoting
		mock.ExpectQuery(`SELECT "quotation_id" FROM "process_quotations" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"quotation_id"}))

		mock.ExpectQuery(`SELECT "sales_order_id" FROM "process_sales_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"sales_order_id"}))

		mock.ExpectQuery(`SELECT "purchase_order_id" FROM "process_purchase_orders" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"purchase_order_id"}))

		mock.ExpectQuery(`SELECT "delivery_id" FROM "process_deliveries" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"delivery_id"}))

		mock.ExpectQuery(`SELECT "invoice_id" FROM "process_invoices" WHERE process_id = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{"invoice_id"}))
	}

	// Execute function to test
	result, err := repo.GetSalesProcessByStatus(status, nil)
	if err != nil {
		t.Fatalf("Error getting sales processes by status: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify processes were returned with correct status
	processes, ok := result.Items.([]models.SalesProcess)
	if !ok {
		t.Fatalf("Could not convert items to []models.SalesProcess")
	}

	for i, p := range processes {
		if p.Status != status {
			t.Errorf("Process %d has incorrect status. Expected: %s, Got: %s",
				i, status, p.Status)
		}
	}
}
