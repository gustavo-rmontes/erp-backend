package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	repository "ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// createMockPaymentRepo creates a repository with mocked DB
func createMockPaymentRepo(db *gorm.DB) repository.PaymentRepository {
	return repository.NewTestPaymentRepository(db)
}

// createTestPayment creates a test payment with given ID
func createTestPayment(id int) *models.Payment {
	return &models.Payment{
		ID:            id,
		InvoiceID:     1,
		Amount:        100.00,
		PaymentDate:   time.Now(),
		PaymentMethod: "credit_card",
		Reference:     "REF-001",
		Notes:         "Pagamento de teste",
		Invoice: &models.Invoice{
			ID:           1,
			InvoiceNo:    "INV-001",
			ContactID:    1,
			Status:       models.InvoiceStatusSent,
			GrandTotal:   200.00,
			AmountPaid:   100.00,
			DueDate:      time.Now().AddDate(0, 1, 0),
			PaymentTerms: "Net 30",
		},
	}
}

// TestCreatePayment tests creating a new payment
func TestCreatePayment(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Create test payment
	payment := createTestPayment(0) // ID 0 because it will be assigned by the DB

	// Remove related objects to avoid GORM trying to save them
	originalInvoice := payment.Invoice
	payment.Invoice = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect payment insert
	mock.ExpectQuery(`INSERT INTO "payments"`).
		WithArgs(
			sqlmock.AnyArg(), // invoice_id
			sqlmock.AnyArg(), // amount
			sqlmock.AnyArg(), // payment_date
			sqlmock.AnyArg(), // payment_method
			sqlmock.AnyArg(), // reference
			sqlmock.AnyArg(), // notes
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect invoice fetch for updating amount paid
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(payment.InvoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "contact_id", "status", "grand_total", "amount_paid",
			"created_at", "updated_at", "due_date", "payment_terms",
		}).AddRow(
			originalInvoice.ID, originalInvoice.InvoiceNo, originalInvoice.ContactID,
			originalInvoice.Status, originalInvoice.GrandTotal, originalInvoice.AmountPaid,
			time.Now(), time.Now(), originalInvoice.DueDate, originalInvoice.PaymentTerms,
		))

	// Use ExpectExec with AnyArg() for SQL update
	mock.ExpectExec(`UPDATE "invoices" SET`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreatePayment(payment); err != nil {
		t.Fatalf("Error creating payment: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify payment ID was set
	if payment.ID != 1 {
		t.Errorf("Expected payment ID to be 1, got %d", payment.ID)
	}
}

// TestGetPaymentByID tests retrieving a payment by ID
func TestGetPaymentByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Setup test data
	testPayment := createTestPayment(1)
	paymentID := testPayment.ID

	// Setup expectations for finding payment
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(paymentID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
		}).AddRow(
			testPayment.ID, testPayment.InvoiceID, testPayment.Amount,
			testPayment.PaymentDate, testPayment.PaymentMethod, testPayment.Reference, testPayment.Notes,
		))

	// Setup expectations for loading invoice
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1`).
		WithArgs(testPayment.InvoiceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "contact_id", "status", "grand_total", "amount_paid",
		}).AddRow(
			testPayment.Invoice.ID, testPayment.Invoice.InvoiceNo,
			testPayment.Invoice.ContactID, testPayment.Invoice.Status,
			testPayment.Invoice.GrandTotal, testPayment.Invoice.AmountPaid,
		))

	// Execute function to test
	retrievedPayment, err := repo.GetPaymentByID(paymentID)
	if err != nil {
		t.Fatalf("Error retrieving payment: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved payment data
	if retrievedPayment.ID != paymentID {
		t.Errorf("Expected payment ID %d, got %d", paymentID, retrievedPayment.ID)
	}

	if retrievedPayment.Amount != testPayment.Amount {
		t.Errorf("Expected amount %.2f, got %.2f", testPayment.Amount, retrievedPayment.Amount)
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetPaymentByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAllPayments tests retrieving all payments with pagination
func TestGetAllPayments(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Test data
	totalItems := int64(3)
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payments"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
	})

	// Add test payments to results
	for i := 1; i <= int(totalItems); i++ {
		testPay := createTestPayment(i)
		rows.AddRow(
			testPay.ID, testPay.InvoiceID, testPay.Amount,
			testPay.PaymentDate, testPay.PaymentMethod, testPay.Reference, testPay.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "payments" ORDER BY payment_date DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading invoices for each payment
	for i := 1; i <= int(totalItems); i++ {
		mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1`).
			WithArgs(1). // All test payments have invoice_id = 1
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_no", "contact_id", "status",
			}).AddRow(
				1, "INV-001", 1, "sent",
			))
	}

	// Execute function to test
	result, err := repo.GetAllPayments(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all payments: %v", err)
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

	// Verify payments were returned
	payments, ok := result.Items.([]models.Payment)
	if !ok {
		t.Fatalf("Could not convert items to []models.Payment")
	}

	if len(payments) != int(totalItems) {
		t.Errorf("Expected %d payments, got %d", totalItems, len(payments))
	}

	// Sub-test for invalid pagination parameters
	t.Run("Invalid pagination", func(t *testing.T) {
		// Setup invalid pagination params
		invalidParams := &pagination.PaginationParams{
			Page:     0, // Invalid page (should be >= 1)
			PageSize: 10,
		}

		// Execute function to test with invalid pagination
		result, err := repo.GetAllPayments(invalidParams)

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

// TestUpdatePayment tests updating an existing payment
func TestUpdatePayment(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Test data
	paymentID := 1
	initialPayment := createTestPayment(paymentID)

	// Create updated payment data with higher amount
	updatedPayment := &models.Payment{
		ID:            paymentID,
		InvoiceID:     initialPayment.InvoiceID,
		Amount:        150.00, // Increased from initial 100.00
		PaymentDate:   time.Now(),
		PaymentMethod: "bank_transfer", // Changed from initial credit_card
		Reference:     "REF-002",
		Notes:         "Pagamento atualizado",
	}

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect query to check if payment exists and get current data
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(paymentID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
		}).AddRow(
			initialPayment.ID, initialPayment.InvoiceID, initialPayment.Amount,
			initialPayment.PaymentDate, initialPayment.PaymentMethod, initialPayment.Reference, initialPayment.Notes,
		))

	// Expect invoice fetch to update amount paid
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(initialPayment.InvoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "contact_id", "status", "grand_total", "amount_paid",
			"created_at", "updated_at", "due_date",
		}).AddRow(
			initialPayment.Invoice.ID, initialPayment.Invoice.InvoiceNo, initialPayment.Invoice.ContactID,
			initialPayment.Invoice.Status, initialPayment.Invoice.GrandTotal, initialPayment.Invoice.AmountPaid,
			time.Now(), time.Now(), initialPayment.Invoice.DueDate,
		))

	// Use ExpectExec with AnyArg() for invoice SQL update - matching all fields GORM updates
	mock.ExpectExec(`UPDATE "invoices" SET`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect payment update - fixed to expect 8 arguments
	mock.ExpectExec(`UPDATE "payments" SET`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdatePayment(paymentID, updatedPayment)
	if err != nil {
		t.Fatalf("Error updating payment: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.UpdatePayment(invalidID, updatedPayment)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestDeletePayment tests deleting a payment
func TestDeletePayment(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Test data
	paymentID := 1
	testPayment := createTestPayment(paymentID)

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect query to get payment data before deletion
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(paymentID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
		}).AddRow(
			testPayment.ID, testPayment.InvoiceID, testPayment.Amount,
			testPayment.PaymentDate, testPayment.PaymentMethod, testPayment.Reference, testPayment.Notes,
		))

	// Expect invoice fetch to update amount paid
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(testPayment.InvoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "contact_id", "status", "grand_total", "amount_paid",
			"created_at", "updated_at", "due_date",
		}).AddRow(
			testPayment.Invoice.ID, testPayment.Invoice.InvoiceNo, testPayment.Invoice.ContactID,
			testPayment.Invoice.Status, testPayment.Invoice.GrandTotal, testPayment.Invoice.AmountPaid,
			time.Now(), time.Now(), testPayment.Invoice.DueDate,
		))

	// Use ExpectExec with AnyArg() for invoice SQL update
	mock.ExpectExec(`UPDATE "invoices" SET`).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect payment deletion
	mock.ExpectExec(`DELETE FROM "payments" WHERE "payments"."id" = \$1`).
		WithArgs(paymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeletePayment(paymentID)
	if err != nil {
		t.Fatalf("Error deleting payment: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."id" = \$1 ORDER BY "payments"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.DeletePayment(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPaymentsByInvoice tests retrieving payments by invoice ID
func TestGetPaymentsByInvoice(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Test data
	invoiceID := 1
	paymentCount := 2

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
	})

	// Add test payments for the invoice
	for i := 1; i <= paymentCount; i++ {
		testPay := createTestPayment(i)
		testPay.InvoiceID = invoiceID
		rows.AddRow(
			testPay.ID, testPay.InvoiceID, testPay.Amount,
			testPay.PaymentDate, testPay.PaymentMethod, testPay.Reference, testPay.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE invoice_id = \$1 ORDER BY payment_date DESC`).
		WithArgs(invoiceID).
		WillReturnRows(rows)

	// Execute function to test
	payments, err := repo.GetPaymentsByInvoice(invoiceID)
	if err != nil {
		t.Fatalf("Error getting payments by invoice: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify payments count
	if len(payments) != paymentCount {
		t.Errorf("Expected %d payments, got %d", paymentCount, len(payments))
	}

	// Verify all payments have the correct invoice ID
	for i, p := range payments {
		if p.InvoiceID != invoiceID {
			t.Errorf("Payment %d has incorrect invoice ID. Expected: %d, Got: %d",
				i, invoiceID, p.InvoiceID)
		}
	}
}

// TestGetPaymentsByDateRange tests retrieving payments within a date range
func TestGetPaymentsByDateRange(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPaymentRepo(db)

	// Test data
	startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	endDate := time.Now()
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payments" WHERE payment_date BETWEEN \$1 AND \$2`).
		WithArgs(startDate, sqlmock.AnyArg()). // End date is adjusted in the code
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
	})

	// Add test payments to results
	for i := 1; i <= int(totalItems); i++ {
		testPay := createTestPayment(i)
		// Set payment_date within the range
		testPay.PaymentDate = startDate.AddDate(0, 0, i) // Spread over the month
		rows.AddRow(
			testPay.ID, testPay.InvoiceID, testPay.Amount,
			testPay.PaymentDate, testPay.PaymentMethod, testPay.Reference, testPay.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE payment_date BETWEEN \$1 AND \$2 ORDER BY payment_date DESC LIMIT \$3`).
		WithArgs(startDate, sqlmock.AnyArg(), pageSize).
		WillReturnRows(rows)

	// Expect queries for loading invoices for each payment
	for i := 1; i <= int(totalItems); i++ {
		mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1`).
			WithArgs(1). // All test payments have invoice_id = 1
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_no", "contact_id", "status",
			}).AddRow(
				1, "INV-001", 1, "sent",
			))
	}

	// Execute function to test
	result, err := repo.GetPaymentsByDateRange(startDate, endDate, nil)
	if err != nil {
		t.Fatalf("Error getting payments by date range: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify pagination result data
	if result.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, result.TotalItems)
	}

	// Verify payments were returned
	payments, ok := result.Items.([]models.Payment)
	if !ok {
		t.Fatalf("Could not convert items to []models.Payment")
	}

	if len(payments) != int(totalItems) {
		t.Errorf("Expected %d payments, got %d", totalItems, len(payments))
	}
}
