package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	repository "ERP-ONSMART/backend/internal/modules/sales/repository"

	"ERP-ONSMART/backend/internal/utils/pagination"

	"fmt"
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

// createMockSalesOrderRepo creates a repository with mocked DB
func createMockSalesOrderRepo(db *gorm.DB) repository.SalesOrderRepository {
	return repository.NewTestSalesOrderRepository(db)
}

// createTestSalesOrder creates a test sales order with given ID
func createTestSalesOrder(id int) *models.SalesOrder {
	return &models.SalesOrder{
		ID:              id,
		SONo:            fmt.Sprintf("TEST-SO-%03d", id),
		ContactID:       1,
		QuotationID:     1,
		Status:          models.SOStatusDraft,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpectedDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:        1000.00,
		TaxTotal:        150.00,
		DiscountTotal:   50.00,
		GrandTotal:      1100.00,
		Notes:           "Pedido de venda de teste",
		PaymentTerms:    "Net 30",
		ShippingAddress: "Rua Teste, 123 - São Paulo/SP",
		Items: []models.SOItem{
			{
				ID:           1,
				SalesOrderID: id,
				ProductID:    1,
				ProductName:  "Produto de Teste",
				ProductCode:  "PROD-001",
				Description:  "Descrição do produto de teste",
				Quantity:     10,
				UnitPrice:    100.00,
				Discount:     5.00,
				Tax:          15.00,
				Total:        1100.00,
			},
		},
		Contact: &contact.Contact{
			ID:         1,
			Name:       "Test Client",
			PersonType: "pj",
			Type:       "cliente",
			Document:   "12345678901234",
			Email:      "client@example.com",
			ZipCode:    "12345-678",
		},
		Quotation: &models.Quotation{
			ID:            1,
			QuotationNo:   "TEST-QUOTE-001",
			ContactID:     1,
			Status:        models.QuotationStatusAccepted,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			ExpiryDate:    time.Now().AddDate(0, 1, 0),
			SubTotal:      1000.00,
			TaxTotal:      150.00,
			DiscountTotal: 50.00,
			GrandTotal:    1100.00,
			Notes:         "Cotação de teste",
			Terms:         "Termos e condições padrão",
		},
	}
}

// TestCreateSalesOrder tests the creation of a new sales order
func TestCreateSalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Create test sales order
	order := createTestSalesOrder(0) // ID 0 because it will be assigned by the DB

	// Remove related objects to avoid GORM trying to save them
	order.Contact = nil
	order.Quotation = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect order insert - using ExpectQuery to capture the RETURNING id
	mock.ExpectQuery(`INSERT INTO "sales_orders"`).
		WithArgs(
			sqlmock.AnyArg(), // so_no
			sqlmock.AnyArg(), // quotation_id
			sqlmock.AnyArg(), // contact_id
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // expected_date
			sqlmock.AnyArg(), // subtotal
			sqlmock.AnyArg(), // tax_total
			sqlmock.AnyArg(), // discount_total
			sqlmock.AnyArg(), // grand_total
			sqlmock.AnyArg(), // notes
			sqlmock.AnyArg(), // payment_terms
			sqlmock.AnyArg(), // shipping_address
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect item insert
	for range order.Items {
		mock.ExpectQuery(`INSERT INTO "sales_order_items"`).
			WithArgs(
				sqlmock.AnyArg(), // sales_order_id
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
	mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE sales_order_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sales_order_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			1, 1, 1, "Produto de Teste", "PROD-001",
			"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
		))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreateSalesOrder(order); err != nil {
		t.Fatalf("Error creating sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify sales order ID was set
	if order.ID != 1 {
		t.Errorf("Expected sales order ID to be 1, got %d", order.ID)
	}
}

// TestGetSalesOrderByID tests retrieving a sales order by ID
func TestGetSalesOrderByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Setup test data
	testOrder := createTestSalesOrder(1)
	orderID := testOrder.ID

	// Setup expectations for finding sales order
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(orderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "quotation_id", "contact_id", "status",
			"created_at", "updated_at", "expected_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "payment_terms", "shipping_address",
		}).AddRow(
			testOrder.ID, testOrder.SONo, testOrder.QuotationID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		))

	// Setup expectations for loading items
	mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE "sales_order_items"."sales_order_id" = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sales_order_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			testOrder.Items[0].ID, testOrder.Items[0].SalesOrderID,
			testOrder.Items[0].ProductID, testOrder.Items[0].ProductName,
			testOrder.Items[0].ProductCode, testOrder.Items[0].Description,
			testOrder.Items[0].Quantity, testOrder.Items[0].UnitPrice,
			testOrder.Items[0].Discount, testOrder.Items[0].Tax,
			testOrder.Items[0].Total,
		))

	// Setup expectations for loading contact
	mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
		WithArgs(testOrder.ContactID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "person_type", "type", "name", "document", "email", "zip_code",
		}).AddRow(
			testOrder.Contact.ID,
			testOrder.Contact.PersonType,
			testOrder.Contact.Type,
			testOrder.Contact.Name,
			testOrder.Contact.Document,
			testOrder.Contact.Email,
			testOrder.Contact.ZipCode,
		))

	// Setup expectations for loading quotation
	mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
		WithArgs(testOrder.QuotationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_no", "contact_id", "status",
			"created_at", "updated_at", "expiry_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "terms",
		}).AddRow(
			testOrder.Quotation.ID, testOrder.Quotation.QuotationNo,
			testOrder.Quotation.ContactID, testOrder.Quotation.Status,
			testOrder.Quotation.CreatedAt, testOrder.Quotation.UpdatedAt,
			testOrder.Quotation.ExpiryDate,
			testOrder.Quotation.SubTotal, testOrder.Quotation.TaxTotal,
			testOrder.Quotation.DiscountTotal, testOrder.Quotation.GrandTotal,
			testOrder.Quotation.Notes, testOrder.Quotation.Terms,
		))

	// Execute function to test
	retrievedOrder, err := repo.GetSalesOrderByID(orderID)
	if err != nil {
		t.Fatalf("Error retrieving sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved sales order data
	if retrievedOrder.ID != orderID {
		t.Errorf("Expected sales order ID %d, got %d", orderID, retrievedOrder.ID)
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetSalesOrderByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}
}

// TestGetAllSalesOrders tests retrieving all sales orders with pagination
func TestGetAllSalesOrders(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_orders"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "so_no", "quotation_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test sales orders to results
	for i := 1; i <= int(totalItems); i++ {
		testOrder := createTestSalesOrder(i)
		rows.AddRow(
			testOrder.ID, testOrder.SONo, testOrder.QuotationID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items, contacts, and quotations for each sales order
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE "sales_order_items"."sales_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "sales_order_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "person_type", "type", "document", "email", "zip_code",
			}).AddRow(
				1, "Test Client", "pj", "cliente", "12345678901234", "client@example.com", "12345-678",
			))

		// Quotation query
		mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_no", "contact_id", "status",
				"created_at", "updated_at", "expiry_date",
				"subtotal", "tax_total", "discount_total", "grand_total",
				"notes", "terms",
			}).AddRow(
				1, "TEST-QUOTE-001", 1, "accepted",
				time.Now(), time.Now(), time.Now().AddDate(0, 1, 0),
				1000.00, 150.00, 50.00, 1100.00,
				"Cotação de teste", "Termos e condições padrão",
			))
	}

	// Execute function to test
	result, err := repo.GetAllSalesOrders(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all sales orders: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify pagination result data
	if result.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, result.TotalItems)
	}

	expectedTotalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if result.TotalPages != expectedTotalPages {
		t.Errorf("Expected total pages %d, got %d", expectedTotalPages, result.TotalPages)
	}

	// Test with invalid pagination parameters
	t.Run("Invalid pagination", func(t *testing.T) {
		// Setup invalid pagination params
		invalidParams := &pagination.PaginationParams{
			Page:     0, // Invalid page (should be >= 1)
			PageSize: 10,
		}

		// Execute function with invalid pagination
		result, err := repo.GetAllSalesOrders(invalidParams)

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

// TestUpdateSalesOrder tests updating an existing sales order
func TestUpdateSalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	orderID := 1
	initialOrder := createTestSalesOrder(orderID)

	// Create updated sales order data
	updatedOrder := &models.SalesOrder{
		ID:              orderID,
		SONo:            initialOrder.SONo,
		QuotationID:     initialOrder.QuotationID,
		ContactID:       initialOrder.ContactID,
		Status:          models.SOStatusConfirmed, // Updated status
		CreatedAt:       initialOrder.CreatedAt,
		UpdatedAt:       time.Now(),
		ExpectedDate:    time.Now().AddDate(0, 2, 0), // Updated date
		SubTotal:        2000.00,                     // Updated values
		TaxTotal:        300.00,
		DiscountTotal:   100.00,
		GrandTotal:      2200.00,
		Notes:           "Pedido atualizado para teste",
		PaymentTerms:    "Net 45",
		ShippingAddress: "Rua Atualizada, 456 - São Paulo/SP",
		Items: []models.SOItem{
			{
				ProductID:   1,
				ProductName: "Produto Atualizado",
				ProductCode: "PROD-001",
				Description: "Descrição atualizada",
				Quantity:    20,
				UnitPrice:   110.00,
				Discount:    10.00,
				Tax:         20.00,
				Total:       2200.00,
			},
			{
				// New item
				ProductID:   2,
				ProductName: "Produto Adicional",
				ProductCode: "PROD-002",
				Description: "Descrição adicional",
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

	// Expect query to check if sales order exists
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(orderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "quotation_id", "contact_id", "status",
			"created_at", "updated_at", "expected_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "payment_terms", "shipping_address",
		}).AddRow(
			initialOrder.ID, initialOrder.SONo, initialOrder.QuotationID,
			initialOrder.ContactID, initialOrder.Status, initialOrder.CreatedAt, initialOrder.UpdatedAt,
			initialOrder.ExpectedDate, initialOrder.SubTotal, initialOrder.TaxTotal,
			initialOrder.DiscountTotal, initialOrder.GrandTotal, initialOrder.Notes,
			initialOrder.PaymentTerms, initialOrder.ShippingAddress,
		))

	// Expect sales order update
	mock.ExpectExec(`UPDATE "sales_orders" SET`).
		WithArgs(
			orderID,
			initialOrder.SONo,
			initialOrder.QuotationID,
			initialOrder.ContactID,
			models.SOStatusConfirmed,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			updatedOrder.ExpectedDate,
			updatedOrder.SubTotal,
			updatedOrder.TaxTotal,
			updatedOrder.DiscountTotal,
			updatedOrder.GrandTotal,
			updatedOrder.Notes,
			updatedOrder.PaymentTerms,
			updatedOrder.ShippingAddress,
			orderID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect deletion of existing items
	mock.ExpectExec(`DELETE FROM "sales_order_items" WHERE sales_order_id = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect batch insert of new items
	mock.ExpectQuery(`INSERT INTO "sales_order_items" \("sales_order_id","product_id","product_name","product_code","description","quantity","unit_price","discount","tax","total"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\),\(\$11,\$12,\$13,\$14,\$15,\$16,\$17,\$18,\$19,\$20\) RETURNING "id"`).
		WithArgs(
			// First item
			orderID,
			updatedOrder.Items[0].ProductID,
			updatedOrder.Items[0].ProductName,
			updatedOrder.Items[0].ProductCode,
			updatedOrder.Items[0].Description,
			updatedOrder.Items[0].Quantity,
			updatedOrder.Items[0].UnitPrice,
			updatedOrder.Items[0].Discount,
			updatedOrder.Items[0].Tax,
			updatedOrder.Items[0].Total,
			// Second item
			orderID,
			updatedOrder.Items[1].ProductID,
			updatedOrder.Items[1].ProductName,
			updatedOrder.Items[1].ProductCode,
			updatedOrder.Items[1].Description,
			updatedOrder.Items[1].Quantity,
			updatedOrder.Items[1].UnitPrice,
			updatedOrder.Items[1].Discount,
			updatedOrder.Items[1].Tax,
			updatedOrder.Items[1].Total,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdateSalesOrder(orderID, updatedOrder)
	if err != nil {
		t.Fatalf("Error updating sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	t.Run("Invalid ID", func(t *testing.T) {
		invalidID := -1
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
			WithArgs(invalidID, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		err = repo.UpdateSalesOrder(invalidID, updatedOrder)
		if err == nil {
			t.Errorf("Expected error for invalid ID, got nil")
		}
	})
}

// TestDeleteSalesOrder tests deleting a sales order by ID
func TestDeleteSalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	orderID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect check for related purchase orders
	mock.ExpectQuery(`SELECT count\(\*\) FROM "purchase_orders" WHERE sales_order_id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect check for related invoices
	mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE sales_order_id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect deletion of items
	mock.ExpectExec(`DELETE FROM "sales_order_items" WHERE sales_order_id = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of sales order
	mock.ExpectExec(`DELETE FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeleteSalesOrder(orderID)
	if err != nil {
		t.Fatalf("Error deleting sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with existing related records (purchase orders)
	t.Run("With related PO records", func(t *testing.T) {
		orderID = 2
		mock.ExpectBegin()

		// Expect check for related purchase orders - found 1 record
		mock.ExpectQuery(`SELECT count\(\*\) FROM "purchase_orders" WHERE sales_order_id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Expect rollback since related records exist
		mock.ExpectRollback()

		// Execute function to test
		err = repo.DeleteSalesOrder(orderID)
		if err == nil {
			t.Errorf("Expected error for sales order with related purchase orders, got nil")
		}

		// Verify error is ErrRelatedRecordsExist
		expectedErr := fmt.Errorf("%w: pedido possui %d pedidos de compra vinculados", errors.ErrRelatedRecordsExist, 1)
		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	// Test with existing related records (invoices)
	t.Run("With related invoice records", func(t *testing.T) {
		orderID = 3
		mock.ExpectBegin()

		// Expect check for related purchase orders - found 0 records
		mock.ExpectQuery(`SELECT count\(\*\) FROM "purchase_orders" WHERE sales_order_id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Expect check for related invoices - found 1 record
		mock.ExpectQuery(`SELECT count\(\*\) FROM "invoices" WHERE sales_order_id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Expect rollback since related records exist
		mock.ExpectRollback()

		// Execute function to test
		err = repo.DeleteSalesOrder(orderID)
		if err == nil {
			t.Errorf("Expected error for sales order with related invoices, got nil")
		}

		// Verify error is ErrRelatedRecordsExist
		expectedErr := fmt.Errorf("%w: pedido possui %d faturas vinculadas", errors.ErrRelatedRecordsExist, 1)
		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestGetSalesOrdersByStatus tests retrieving sales orders by status
func TestGetSalesOrdersByStatus(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	status := models.SOStatusConfirmed
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_orders" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "so_no", "quotation_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test sales orders to results
	for i := 1; i <= int(totalItems); i++ {
		testOrder := createTestSalesOrder(i)
		testOrder.Status = status
		rows.AddRow(
			testOrder.ID, testOrder.SONo, testOrder.QuotationID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern with OrderBy delivery_date ASC
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE status = \$1 ORDER BY delivery_date ASC LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each sales order
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE "sales_order_items"."sales_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "sales_order_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "type", "email",
			}).AddRow(
				1, "Test Client", "cliente", "client@example.com",
			))

		// Quotation query
		mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_no", "contact_id", "status",
				"created_at", "updated_at", "expiry_date",
				"subtotal", "tax_total", "discount_total", "grand_total",
				"notes", "terms",
			}).AddRow(
				1, "TEST-QUOTE-001", 1, "accepted",
				time.Now(), time.Now(), time.Now().AddDate(0, 1, 0),
				1000.00, 150.00, 50.00, 1100.00,
				"Cotação de teste", "Termos e condições padrão",
			))
	}

	// Execute function to test
	result, err := repo.GetSalesOrdersByStatus(status, nil)
	if err != nil {
		t.Fatalf("Error getting sales orders by status: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify sales orders were returned with correct status
	orders, ok := result.Items.([]models.SalesOrder)
	if !ok {
		t.Fatalf("Could not convert items to []models.SalesOrder")
	}

	for i, o := range orders {
		if o.Status != status {
			t.Errorf("Sales order %d has incorrect status. Expected: %s, Got: %s",
				i, status, o.Status)
		}
	}
}

// TestGetSalesOrdersByContact tests retrieving sales orders by contact ID
func TestGetSalesOrdersByContact(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	contactID := 1
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sales_orders" WHERE contact_id = \$1`).
		WithArgs(contactID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "so_no", "quotation_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test sales orders to results
	for i := 1; i <= int(totalItems); i++ {
		testOrder := createTestSalesOrder(i)
		rows.AddRow(
			testOrder.ID, testOrder.SONo, testOrder.QuotationID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern - Changed to match what GORM actually generates
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE contact_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(contactID, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each sales order
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE "sales_order_items"."sales_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "sales_order_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(contactID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "type", "email",
			}).AddRow(
				contactID, "Test Client", "cliente", "client@example.com",
			))

		// Quotation query
		mock.ExpectQuery(`SELECT \* FROM "quotations" WHERE "quotations"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "quotation_no", "contact_id", "status",
				"created_at", "updated_at", "expiry_date",
				"subtotal", "tax_total", "discount_total", "grand_total",
				"notes", "terms",
			}).AddRow(
				1, "TEST-QUOTE-001", contactID, "accepted",
				time.Now(), time.Now(), time.Now().AddDate(0, 1, 0),
				1000.00, 150.00, 50.00, 1100.00,
				"Cotação de teste", "Termos e condições padrão",
			))
	}

	// Execute function to test
	result, err := repo.GetSalesOrdersByContact(contactID, nil)
	if err != nil {
		t.Fatalf("Error getting sales orders by contact: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify sales orders were returned with correct contact ID
	orders, ok := result.Items.([]models.SalesOrder)
	if !ok {
		t.Fatalf("Could not convert items to []models.SalesOrder")
	}

	for i, o := range orders {
		if o.ContactID != contactID {
			t.Errorf("Sales order %d has incorrect contact ID. Expected: %d, Got: %d",
				i, contactID, o.ContactID)
		}
	}
}

// TestGetSalesOrdersByQuotation tests retrieving a sales order by quotation ID
func TestGetSalesOrdersByQuotation(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockSalesOrderRepo(db)

	// Test data
	quotationID := 1
	testOrder := createTestSalesOrder(1)
	testOrder.QuotationID = quotationID

	// Setup expectations for finding sales order by quotation ID
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE quotation_id = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(quotationID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "quotation_id", "contact_id", "status",
			"created_at", "updated_at", "expected_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "payment_terms", "shipping_address",
		}).AddRow(
			testOrder.ID, testOrder.SONo, testOrder.QuotationID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		))

	// Setup expectations for loading items
	mock.ExpectQuery(`SELECT \* FROM "sales_order_items" WHERE "sales_order_items"."sales_order_id" = \$1`).
		WithArgs(testOrder.ID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sales_order_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			testOrder.Items[0].ID, testOrder.Items[0].SalesOrderID,
			testOrder.Items[0].ProductID, testOrder.Items[0].ProductName,
			testOrder.Items[0].ProductCode, testOrder.Items[0].Description,
			testOrder.Items[0].Quantity, testOrder.Items[0].UnitPrice,
			testOrder.Items[0].Discount, testOrder.Items[0].Tax,
			testOrder.Items[0].Total,
		))

	// Execute function to test
	retrievedOrder, err := repo.GetSalesOrdersByQuotation(quotationID)
	if err != nil {
		t.Fatalf("Error retrieving sales order by quotation: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved order data
	if retrievedOrder.QuotationID != quotationID {
		t.Errorf("Expected quotation ID %d, got %d", quotationID, retrievedOrder.QuotationID)
	}

	// Test non-existent quotation
	nonExistentID := 999
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE quotation_id = \$1 ORDER BY "sales_orders"."id" LIMIT \$2`).
		WithArgs(nonExistentID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetSalesOrdersByQuotation(nonExistentID)
	if err != nil {
		// Should not return an error, just nil
		t.Errorf("Expected nil error for non-existent quotation, got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result for non-existent quotation, got non-nil result")
	}
}
