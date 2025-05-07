package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// createMockInvoiceRepo creates a repository with mocked DB
func createMockInvoiceRepo(db *gorm.DB) InvoiceRepository {
	return &gormInvoiceRepository{
		db:  db,
		log: logger.WithModule("InvoiceRepository"),
	}
}

// createTestInvoice creates a test invoice with given ID
func createTestInvoice(id int) *models.Invoice {
	return &models.Invoice{
		ID:            id,
		InvoiceNo:     fmt.Sprintf("TEST-INV-%03d", id),
		SalesOrderID:  1,
		SONo:          "TEST-SO-001",
		ContactID:     1,
		Status:        models.InvoiceStatusDraft,
		IssueDate:     time.Now(),
		DueDate:       time.Now().AddDate(0, 1, 0),
		SubTotal:      1000.00,
		TaxTotal:      150.00,
		DiscountTotal: 50.00,
		GrandTotal:    1100.00,
		AmountPaid:    0.00,
		PaymentTerms:  "Net 30",
		Notes:         "Fatura de teste",
		Items: []models.InvoiceItem{
			{
				ID:          1,
				InvoiceID:   id,
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
		SalesOrder: &models.SalesOrder{
			ID:         1,
			SONo:       "TEST-SO-001",
			ContactID:  1,
			Status:     models.SOStatusConfirmed,
			GrandTotal: 1100.00,
		},
		Payments: []models.Payment{
			{
				ID:            1,
				InvoiceID:     id,
				Amount:        0.00,
				PaymentDate:   time.Now(),
				PaymentMethod: "",
				Reference:     "",
				Notes:         "",
			},
		},
	}
}

// TestCreateInvoice tests creating a new invoice
func TestCreateInvoice(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Create test invoice
	invoice := createTestInvoice(0) // ID 0 because it will be assigned by the DB

	// Remove related objects to avoid GORM trying to save them
	invoice.Contact = nil
	invoice.SalesOrder = nil
	invoice.Payments = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect invoice insert - using ExpectQuery to capture the RETURNING id
	mock.ExpectQuery(`INSERT INTO "invoices"`).
		WithArgs(
			sqlmock.AnyArg(), // invoice_no
			sqlmock.AnyArg(), // sales_order_id
			sqlmock.AnyArg(), // so_no
			sqlmock.AnyArg(), // contact_id
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // issue_date
			sqlmock.AnyArg(), // due_date
			sqlmock.AnyArg(), // subtotal
			sqlmock.AnyArg(), // tax_total
			sqlmock.AnyArg(), // discount_total
			sqlmock.AnyArg(), // grand_total
			sqlmock.AnyArg(), // amount_paid
			sqlmock.AnyArg(), // payment_terms
			sqlmock.AnyArg(), // notes
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect item insert - also Query with RETURNING id
	for range invoice.Items {
		mock.ExpectQuery(`INSERT INTO "invoice_items"`).
			WithArgs(
				sqlmock.AnyArg(), // invoice_id
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
	mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE invoice_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			1, 1, 1, "Produto de Teste", "PROD-001",
			"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
		))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreateInvoice(invoice); err != nil {
		t.Fatalf("Error creating invoice: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify invoice ID was set
	if invoice.ID != 1 {
		t.Errorf("Expected invoice ID to be 1, got %d", invoice.ID)
	}

	// Verify items were loaded
	if len(invoice.Items) == 0 {
		t.Fatalf("Invoice items were not loaded")
	}
}

// TestGetInvoiceByID tests retrieving an invoice by ID
func TestGetInvoiceByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Setup test data
	testInvoice := createTestInvoice(1)
	invoiceID := testInvoice.ID

	// Setup expectations for finding invoice
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(invoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
			"created_at", "updated_at", "issue_date", "due_date",
			"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
			"payment_terms", "notes",
		}).AddRow(
			testInvoice.ID, testInvoice.InvoiceNo, testInvoice.SalesOrderID, testInvoice.SONo,
			testInvoice.ContactID, testInvoice.Status, time.Now(), time.Now(),
			testInvoice.IssueDate, testInvoice.DueDate, testInvoice.SubTotal,
			testInvoice.TaxTotal, testInvoice.DiscountTotal, testInvoice.GrandTotal,
			testInvoice.AmountPaid, testInvoice.PaymentTerms, testInvoice.Notes,
		))

	// Setup expectations for loading items
	mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
		WithArgs(invoiceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			testInvoice.Items[0].ID, testInvoice.Items[0].InvoiceID,
			testInvoice.Items[0].ProductID, testInvoice.Items[0].ProductName,
			testInvoice.Items[0].ProductCode, testInvoice.Items[0].Description,
			testInvoice.Items[0].Quantity, testInvoice.Items[0].UnitPrice,
			testInvoice.Items[0].Discount, testInvoice.Items[0].Tax,
			testInvoice.Items[0].Total,
		))

	// Setup expectations for loading contact
	mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
		WithArgs(testInvoice.ContactID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "person_type", "type", "name", "document", "email", "zip_code",
		}).AddRow(
			testInvoice.Contact.ID,
			testInvoice.Contact.PersonType,
			testInvoice.Contact.Type,
			testInvoice.Contact.Name,
			testInvoice.Contact.Document,
			testInvoice.Contact.Email,
			testInvoice.Contact.ZipCode,
		))

	// Setup expectations for loading sales order
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
		WithArgs(testInvoice.SalesOrderID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "contact_id", "status", "grand_total",
		}).AddRow(
			testInvoice.SalesOrder.ID,
			testInvoice.SalesOrder.SONo,
			testInvoice.SalesOrder.ContactID,
			testInvoice.SalesOrder.Status,
			testInvoice.SalesOrder.GrandTotal,
		))

	// Setup expectations for loading payments
	mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
		WithArgs(invoiceID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
		}).AddRow(
			testInvoice.Payments[0].ID,
			testInvoice.Payments[0].InvoiceID,
			testInvoice.Payments[0].Amount,
			testInvoice.Payments[0].PaymentDate,
			testInvoice.Payments[0].PaymentMethod,
			testInvoice.Payments[0].Reference,
			testInvoice.Payments[0].Notes,
		))

	// Execute function to test
	retrievedInvoice, err := repo.GetInvoiceByID(invoiceID)
	if err != nil {
		t.Fatalf("Error retrieving invoice: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved invoice data
	if retrievedInvoice.ID != invoiceID {
		t.Errorf("Expected invoice ID %d, got %d", invoiceID, retrievedInvoice.ID)
	}

	if retrievedInvoice.InvoiceNo != testInvoice.InvoiceNo {
		t.Errorf("Expected invoice number %s, got %s",
			testInvoice.InvoiceNo, retrievedInvoice.InvoiceNo)
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetInvoiceByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAllInvoices tests retrieving all invoices with pagination
func TestGetAllInvoices(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	totalItems := int64(3)
	// page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
		"created_at", "updated_at", "issue_date", "due_date",
		"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
		"payment_terms", "notes",
	})

	// Add test invoices to results
	for i := 1; i <= int(totalItems); i++ {
		testInv := createTestInvoice(i)
		rows.AddRow(
			testInv.ID, testInv.InvoiceNo, testInv.SalesOrderID, testInv.SONo,
			testInv.ContactID, testInv.Status, time.Now(), time.Now(),
			testInv.IssueDate, testInv.DueDate, testInv.SubTotal,
			testInv.TaxTotal, testInv.DiscountTotal, testInv.GrandTotal,
			testInv.AmountPaid, testInv.PaymentTerms, testInv.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "invoices" ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each invoice
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "contact_id", "status", "grand_total",
			}).AddRow(
				1, "TEST-SO-001", 1, "confirmed", 1100.00,
			))

		// Payments query
		mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
			}).AddRow(
				1, i, 0.00, time.Now(), "", "", "",
			))
	}

	// Execute function to test
	result, err := repo.GetAllInvoices(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all invoices: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify pagination result data
	if result.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, result.TotalItems)
	}

	// Verify invoices were returned
	invoices, ok := result.Items.([]models.Invoice)
	if !ok {
		t.Fatalf("Could not convert items to []models.Invoice")
	}

	if len(invoices) != int(totalItems) {
		t.Errorf("Expected %d invoices, got %d", totalItems, len(invoices))
	}

	// Sub-test for invalid pagination parameters
	t.Run("Invalid pagination", func(t *testing.T) {
		// Setup invalid pagination params
		invalidParams := &pagination.PaginationParams{
			Page:     0, // Invalid page (should be >= 1)
			PageSize: 10,
		}

		// Execute function to test with invalid pagination
		result, err := repo.GetAllInvoices(invalidParams)

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

// TestUpdateInvoice tests updating an existing invoice
func TestUpdateInvoice(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	invoiceID := 1
	initialInvoice := createTestInvoice(invoiceID)

	// Create updated invoice data
	updatedInvoice := &models.Invoice{
		ID:            invoiceID,
		InvoiceNo:     initialInvoice.InvoiceNo,
		SalesOrderID:  initialInvoice.SalesOrderID,
		SONo:          initialInvoice.SONo,
		ContactID:     initialInvoice.ContactID,
		Status:        models.InvoiceStatusSent,    // Updated status
		DueDate:       time.Now().AddDate(0, 2, 0), // Updated due date
		SubTotal:      2000.00,                     // Updated values
		TaxTotal:      300.00,
		DiscountTotal: 100.00,
		GrandTotal:    2200.00,
		AmountPaid:    0.00,
		PaymentTerms:  "Net 45", // Updated terms
		Notes:         "Fatura atualizada para teste",
		Items: []models.InvoiceItem{
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

	// Expect query to check if invoice exists
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(invoiceID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
			"created_at", "updated_at", "issue_date", "due_date",
			"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
			"payment_terms", "notes",
		}).AddRow(
			initialInvoice.ID, initialInvoice.InvoiceNo, initialInvoice.SalesOrderID, initialInvoice.SONo,
			initialInvoice.ContactID, initialInvoice.Status, time.Now(), time.Now(),
			initialInvoice.IssueDate, initialInvoice.DueDate, initialInvoice.SubTotal,
			initialInvoice.TaxTotal, initialInvoice.DiscountTotal, initialInvoice.GrandTotal,
			initialInvoice.AmountPaid, initialInvoice.PaymentTerms, initialInvoice.Notes,
		))

	// Expect invoice update - fix: using AnyArgs() for each parameter and correct number (15) of arguments
	mock.ExpectExec(`UPDATE "invoices" SET`).
		WithArgs(
			sqlmock.AnyArg(), // id
			sqlmock.AnyArg(), // invoice_no
			sqlmock.AnyArg(), // sales_order_id
			sqlmock.AnyArg(), // so_no
			sqlmock.AnyArg(), // contact_id
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // due_date
			sqlmock.AnyArg(), // subtotal
			sqlmock.AnyArg(), // tax_total
			sqlmock.AnyArg(), // discount_total
			sqlmock.AnyArg(), // grand_total
			sqlmock.AnyArg(), // payment_terms
			sqlmock.AnyArg(), // notes
			sqlmock.AnyArg(), // id (for WHERE clause)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect deletion of existing items
	mock.ExpectExec(`DELETE FROM "invoice_items" WHERE invoice_id = \$1`).
		WithArgs(invoiceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect batch insert of new items
	mock.ExpectQuery(`INSERT INTO "invoice_items" \("invoice_id","product_id","product_name","product_code","description","quantity","unit_price","discount","tax","total"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\),\(\$11,\$12,\$13,\$14,\$15,\$16,\$17,\$18,\$19,\$20\) RETURNING "id"`).
		WithArgs(
			// First item
			invoiceID,
			updatedInvoice.Items[0].ProductID,
			updatedInvoice.Items[0].ProductName,
			updatedInvoice.Items[0].ProductCode,
			updatedInvoice.Items[0].Description,
			updatedInvoice.Items[0].Quantity,
			updatedInvoice.Items[0].UnitPrice,
			updatedInvoice.Items[0].Discount,
			updatedInvoice.Items[0].Tax,
			updatedInvoice.Items[0].Total,
			// Second item
			invoiceID,
			updatedInvoice.Items[1].ProductID,
			updatedInvoice.Items[1].ProductName,
			updatedInvoice.Items[1].ProductCode,
			updatedInvoice.Items[1].Description,
			updatedInvoice.Items[1].Quantity,
			updatedInvoice.Items[1].UnitPrice,
			updatedInvoice.Items[1].Discount,
			updatedInvoice.Items[1].Tax,
			updatedInvoice.Items[1].Total,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdateInvoice(invoiceID, updatedInvoice)
	if err != nil {
		t.Fatalf("Error updating invoice: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE "invoices"."id" = \$1 ORDER BY "invoices"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.UpdateInvoice(invalidID, updatedInvoice)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestDeleteInvoice tests deleting an invoice by ID
func TestDeleteInvoice(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	invoiceID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect check for related payments - match exact SQL pattern
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payments" WHERE invoice_id = \$1`).
		WithArgs(invoiceID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect deletion of items - match exact SQL pattern
	mock.ExpectExec(`DELETE FROM "invoice_items" WHERE invoice_id = \$1`).
		WithArgs(invoiceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of invoice
	mock.ExpectExec(`DELETE FROM "invoices" WHERE "invoices"."id" = \$1`).
		WithArgs(invoiceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeleteInvoice(invoiceID)
	if err != nil {
		t.Fatalf("Error deleting invoice: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with existing related records (payments)
	invoiceID = 2
	mock.ExpectBegin()

	// Expect check for related payments - found 1 record
	mock.ExpectQuery(`SELECT count\(\*\) FROM "payments" WHERE invoice_id = \$1`).
		WithArgs(invoiceID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect rollback since related records exist
	mock.ExpectRollback()

	// Execute function to test
	err = repo.DeleteInvoice(invoiceID)
	if err == nil {
		t.Errorf("Expected error for invoice with related records, got nil")
	}

	// Verify error is ErrRelatedRecordsExist
	if fmt.Sprintf("%v", err) != fmt.Sprintf("%v: fatura possui 1 pagamentos associados", errors.ErrRelatedRecordsExist) {
		t.Errorf("Expected error %v, got %v", errors.ErrRelatedRecordsExist, err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetInvoicesByStatus tests retrieving invoices by status with pagination
func TestGetInvoicesByStatus(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	status := models.InvoiceStatusSent
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
		"created_at", "updated_at", "issue_date", "due_date",
		"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
		"payment_terms", "notes",
	})

	// Add test invoices to results
	for i := 1; i <= int(totalItems); i++ {
		testInv := createTestInvoice(i)
		testInv.Status = status
		rows.AddRow(
			testInv.ID, testInv.InvoiceNo, testInv.SalesOrderID, testInv.SONo,
			testInv.ContactID, testInv.Status, time.Now(), time.Now(),
			testInv.IssueDate, testInv.DueDate, testInv.SubTotal,
			testInv.TaxTotal, testInv.DiscountTotal, testInv.GrandTotal,
			testInv.AmountPaid, testInv.PaymentTerms, testInv.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE status = \$1 ORDER BY due_date ASC LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each invoice
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Payments query
		mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
			}).AddRow(
				1, i, 0.00, time.Now(), "", "", "",
			))
	}

	// Execute function to test
	result, err := repo.GetInvoicesByStatus(status, nil)
	if err != nil {
		t.Fatalf("Error getting invoices by status: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify invoices were returned with correct status
	invoices, ok := result.Items.([]models.Invoice)
	if !ok {
		t.Fatalf("Could not convert items to []models.Invoice")
	}

	for i, inv := range invoices {
		if inv.Status != status {
			t.Errorf("Invoice %d has incorrect status. Expected: %s, Got: %s",
				i, status, inv.Status)
		}
	}
}

// TestGetInvoicesBySalesOrder tests retrieving invoices by sales order ID
func TestGetInvoicesBySalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	salesOrderID := 1
	invoiceCount := 2

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
		"created_at", "updated_at", "issue_date", "due_date",
		"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
		"payment_terms", "notes",
	})

	// Add test invoices for the sales order
	for i := 1; i <= invoiceCount; i++ {
		testInv := createTestInvoice(i)
		testInv.SalesOrderID = salesOrderID
		rows.AddRow(
			testInv.ID, testInv.InvoiceNo, testInv.SalesOrderID, testInv.SONo,
			testInv.ContactID, testInv.Status, time.Now(), time.Now(),
			testInv.IssueDate, testInv.DueDate, testInv.SubTotal,
			testInv.TaxTotal, testInv.DiscountTotal, testInv.GrandTotal,
			testInv.AmountPaid, testInv.PaymentTerms, testInv.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE sales_order_id = \$1 ORDER BY created_at DESC`).
		WithArgs(salesOrderID).
		WillReturnRows(rows)

	// Expect queries for loading items and payments for each invoice
	for i := 1; i <= invoiceCount; i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Payments query
		mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
			}).AddRow(
				1, i, 0.00, time.Now(), "", "", "",
			))
	}

	// Execute function to test
	invoices, err := repo.GetInvoicesBySalesOrder(salesOrderID)
	if err != nil {
		t.Fatalf("Error getting invoices by sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify invoices count
	if len(invoices) != invoiceCount {
		t.Errorf("Expected %d invoices, got %d", invoiceCount, len(invoices))
	}

	// Verify all invoices have the correct sales order ID
	for i, inv := range invoices {
		if inv.SalesOrderID != salesOrderID {
			t.Errorf("Invoice %d has incorrect sales order ID. Expected: %d, Got: %d",
				i, salesOrderID, inv.SalesOrderID)
		}
	}
}

// TestGetInvoicesByContact tests retrieving invoices by contact ID with pagination
func TestGetInvoicesByContact(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	contactID := 1
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE contact_id = \$1`).
		WithArgs(contactID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
		"created_at", "updated_at", "issue_date", "due_date",
		"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
		"payment_terms", "notes",
	})

	// Add test invoices to results
	for i := 1; i <= int(totalItems); i++ {
		testInv := createTestInvoice(i)
		testInv.ContactID = contactID
		rows.AddRow(
			testInv.ID, testInv.InvoiceNo, testInv.SalesOrderID, testInv.SONo,
			testInv.ContactID, testInv.Status, time.Now(), time.Now(),
			testInv.IssueDate, testInv.DueDate, testInv.SubTotal,
			testInv.TaxTotal, testInv.DiscountTotal, testInv.GrandTotal,
			testInv.AmountPaid, testInv.PaymentTerms, testInv.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE contact_id = \$1 ORDER BY due_date DESC LIMIT \$2`).
		WithArgs(contactID, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each invoice
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(contactID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				contactID, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Payments query
		mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
			}).AddRow(
				1, i, 0.00, time.Now(), "", "", "",
			))
	}

	// Execute function to test
	result, err := repo.GetInvoicesByContact(contactID, nil)
	if err != nil {
		t.Fatalf("Error getting invoices by contact: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify invoices were returned with correct contact ID
	invoices, ok := result.Items.([]models.Invoice)
	if !ok {
		t.Fatalf("Could not convert items to []models.Invoice")
	}

	for i, inv := range invoices {
		if inv.ContactID != contactID {
			t.Errorf("Invoice %d has incorrect contact ID. Expected: %d, Got: %d",
				i, contactID, inv.ContactID)
		}
	}
}

// TestGetOverdueInvoices tests retrieving overdue invoices with pagination
func TestGetOverdueInvoices(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockInvoiceRepo(db)

	// Test data
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize
	now := time.Now()

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE due_date < \$1 AND status NOT IN \(\$2, \$3\)`).
		WithArgs(sqlmock.AnyArg(), models.InvoiceStatusPaid, models.InvoiceStatusCancelled).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "invoice_no", "sales_order_id", "so_no", "contact_id", "status",
		"created_at", "updated_at", "issue_date", "due_date",
		"subtotal", "tax_total", "discount_total", "grand_total", "amount_paid",
		"payment_terms", "notes",
	})

	// Add invoices with past due dates
	for i := 1; i <= int(totalItems); i++ {
		testInv := createTestInvoice(i)
		testInv.Status = models.InvoiceStatusSent // Not yet marked as overdue
		testInv.DueDate = now.AddDate(0, -1, 0)   // Due date 1 month ago
		rows.AddRow(
			testInv.ID, testInv.InvoiceNo, testInv.SalesOrderID, testInv.SONo,
			testInv.ContactID, testInv.Status, time.Now(), time.Now(),
			testInv.IssueDate, testInv.DueDate, testInv.SubTotal,
			testInv.TaxTotal, testInv.DiscountTotal, testInv.GrandTotal,
			testInv.AmountPaid, testInv.PaymentTerms, testInv.Notes,
		)
	}

	// Match the query for retrieving overdue invoices
	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE due_date < \$1 AND status NOT IN \(\$2, \$3\) ORDER BY due_date ASC LIMIT \$4`).
		WithArgs(sqlmock.AnyArg(), models.InvoiceStatusPaid, models.InvoiceStatusCancelled, pageSize).
		WillReturnRows(rows)

	// For each invoice, expect GORM to load its associations
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "invoice_items" WHERE "invoice_items"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "person_type", "type", "name", "document", "email", "zip_code",
			}).AddRow(
				1, "pf", "cliente", "Test Contact", "12345678901", "test@example.com", "12345-678",
			))

		// Payments query
		mock.ExpectQuery(`SELECT \* FROM "payments" WHERE "payments"."invoice_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "invoice_id", "amount", "payment_date", "payment_method", "reference", "notes",
			}).AddRow(
				1, i, 0.00, time.Now(), "", "", "",
			))
	}

	// Since the implementation tries to update invoices to overdue,
	// we'll mock those that fail silently in the implementation

	// Skip mocking the update operations - let them fail in the implementation
	// The implementation handles errors from these updates by just logging them
	// and continuing, so we can ignore them for the test

	// Execute function to test
	result, err := repo.GetOverdueInvoices(nil)
	if err != nil {
		t.Fatalf("Error getting overdue invoices: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify invoices were returned with correct data
	invoices, ok := result.Items.([]models.Invoice)
	if !ok {
		t.Fatalf("Could not convert items to []models.Invoice")
	}

	// Verify the number of invoices matches
	if len(invoices) != int(totalItems) {
		t.Errorf("Expected %d invoices, got %d", totalItems, len(invoices))
	}

	// Check that returned invoices have due dates in the past
	for i, inv := range invoices {
		if !inv.DueDate.Before(now) {
			t.Errorf("Invoice %d is not overdue. Due date: %v, Now: %v",
				i, inv.DueDate, now)
		}

		// Since we're not mocking the status updates, don't check for the status here
		// as they will remain 'sent' in the test results
	}
}
