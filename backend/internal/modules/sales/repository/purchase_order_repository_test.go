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

// createMockPurchaseOrderRepo creates a repository with mocked DB
func createMockPurchaseOrderRepo(db *gorm.DB) PurchaseOrderRepository {
	return &gormPurchaseOrderRepository{
		db:  db,
		log: logger.WithModule("PurchaseOrderRepository"),
	}
}

// createTestPurchaseOrder creates a test purchase order with given ID
func createTestPurchaseOrder(id int) *models.PurchaseOrder {
	return &models.PurchaseOrder{
		ID:              id,
		PONo:            fmt.Sprintf("TEST-PO-%03d", id),
		SONo:            fmt.Sprintf("TEST-SO-%03d", id),
		SalesOrderID:    1,
		ContactID:       1,
		Status:          models.POStatusDraft,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpectedDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:        1000.00,
		TaxTotal:        150.00,
		DiscountTotal:   50.00,
		GrandTotal:      1180.00,
		Notes:           "Pedido de compra de teste",
		PaymentTerms:    "Net 30",
		ShippingAddress: "Rua Teste, 123 - São Paulo/SP",
		Items: []models.POItem{
			{
				ID:              1,
				PurchaseOrderID: id,
				ProductID:       1,
				ProductName:     "Produto de Teste",
				ProductCode:     "PROD-001",
				Description:     "Descrição do produto de teste",
				Quantity:        10,
				UnitPrice:       100.00,
				Discount:        5.00,
				Tax:             15.00,
				Total:           1100.00,
			},
		},
		Contact: &contact.Contact{
			ID:         1,
			Name:       "Test Supplier",
			PersonType: "pj",
			Type:       "fornecedor",
			Document:   "12345678901234",
			Email:      "supplier@example.com",
			ZipCode:    "12345-678",
		},
		SalesOrder: &models.SalesOrder{
			ID:            1,
			SONo:          "TEST-SO-001",
			ContactID:     2,
			Status:        models.SOStatusConfirmed,
			QuotationID:   1,
			ExpectedDate:  time.Now().AddDate(0, 1, 0),
			SubTotal:      1000.00,
			TaxTotal:      150.00,
			DiscountTotal: 50.00,
			GrandTotal:    1100.00,
		},
	}
}

// TestCreatePurchaseOrder tests the creation of a new purchase order
func TestCreatePurchaseOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Create test purchase order
	order := createTestPurchaseOrder(0) // ID 0 because it will be assigned by the DB

	// Remove related objects to avoid GORM trying to save them
	order.Contact = nil
	order.SalesOrder = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect order insert - using ExpectQuery to capture the RETURNING id
	mock.ExpectQuery(`INSERT INTO "purchase_orders"`).
		WithArgs(
			sqlmock.AnyArg(), // po_no
			sqlmock.AnyArg(), // so_no
			sqlmock.AnyArg(), // sales_order_id
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
		mock.ExpectQuery(`INSERT INTO "purchase_order_items"`).
			WithArgs(
				sqlmock.AnyArg(), // purchase_order_id
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
	mock.ExpectQuery(`SELECT \* FROM "purchase_order_items" WHERE purchase_order_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "purchase_order_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			1, 1, 1, "Produto de Teste", "PROD-001",
			"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
		))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreatePurchaseOrder(order); err != nil {
		t.Fatalf("Error creating purchase order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify purchase order ID was set
	if order.ID != 1 {
		t.Errorf("Expected purchase order ID to be 1, got %d", order.ID)
	}
}

// TestGetPurchaseOrderByID tests retrieving a purchase order by ID
func TestGetPurchaseOrderByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Setup test data
	testOrder := createTestPurchaseOrder(1)
	orderID := testOrder.ID

	// Setup expectations for finding purchase order
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1 ORDER BY "purchase_orders"."id" LIMIT \$2`).
		WithArgs(orderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "po_no", "so_no", "sales_order_id", "contact_id", "status",
			"created_at", "updated_at", "expected_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "payment_terms", "shipping_address",
		}).AddRow(
			testOrder.ID, testOrder.PONo, testOrder.SONo, testOrder.SalesOrderID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		))

	// Setup expectations for loading items
	mock.ExpectQuery(`SELECT \* FROM "purchase_order_items" WHERE "purchase_order_items"."purchase_order_id" = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "purchase_order_id", "product_id", "product_name", "product_code",
			"description", "quantity", "unit_price", "discount", "tax", "total",
		}).AddRow(
			testOrder.Items[0].ID, testOrder.Items[0].PurchaseOrderID,
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

	// Setup expectations for loading sales order
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
		WithArgs(testOrder.SalesOrderID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "contact_id", "status", "quotation_id",
			"expected_date", "subtotal", "tax_total", "discount_total", "grand_total",
		}).AddRow(
			testOrder.SalesOrder.ID, testOrder.SalesOrder.SONo,
			testOrder.SalesOrder.ContactID, testOrder.SalesOrder.Status,
			testOrder.SalesOrder.QuotationID, testOrder.SalesOrder.ExpectedDate,
			testOrder.SalesOrder.SubTotal, testOrder.SalesOrder.TaxTotal,
			testOrder.SalesOrder.DiscountTotal, testOrder.SalesOrder.GrandTotal,
		))

	// Execute function to test
	retrievedOrder, err := repo.GetPurchaseOrderByID(orderID)
	if err != nil {
		t.Fatalf("Error retrieving purchase order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved purchase order data
	if retrievedOrder.ID != orderID {
		t.Errorf("Expected purchase order ID %d, got %d", orderID, retrievedOrder.ID)
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1 ORDER BY "purchase_orders"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetPurchaseOrderByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}
}

// TestGetAllPurchaseOrders tests retrieving all purchase orders with pagination
func TestGetAllPurchaseOrders(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Test data
	totalItems := int64(3)
	// page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "purchase_orders"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "po_no", "so_no", "sales_order_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test purchase orders to results
	for i := 1; i <= int(totalItems); i++ {
		testOrder := createTestPurchaseOrder(i)
		rows.AddRow(
			testOrder.ID, testOrder.PONo, testOrder.SONo, testOrder.SalesOrderID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items, contacts, and sales orders for each purchase order
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "purchase_order_items" WHERE "purchase_order_items"."purchase_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "purchase_order_id", "product_id", "product_name", "product_code",
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

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "contact_id", "status",
			}).AddRow(
				1, "TEST-SO-001", 2, "confirmed",
			))
	}

	// Execute function to test
	result, err := repo.GetAllPurchaseOrders(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all purchase orders: %v", err)
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
		result, err := repo.GetAllPurchaseOrders(invalidParams)

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

// TestUpdatePurchaseOrder tests updating an existing purchase order
func TestUpdatePurchaseOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Test data
	orderID := 1
	initialOrder := createTestPurchaseOrder(orderID)

	// Create updated purchase order data
	updatedOrder := &models.PurchaseOrder{
		ID:              orderID,
		PONo:            initialOrder.PONo,
		SONo:            initialOrder.SONo,
		SalesOrderID:    initialOrder.SalesOrderID,
		ContactID:       initialOrder.ContactID,
		Status:          models.POStatusConfirmed, // Updated status
		CreatedAt:       initialOrder.CreatedAt,
		UpdatedAt:       time.Now(),
		ExpectedDate:    time.Now().AddDate(0, 2, 0), // Updated date
		SubTotal:        2000.00,                     // Updated values
		TaxTotal:        300.00,
		DiscountTotal:   100.00,
		GrandTotal:      2320.00,
		Notes:           "Pedido atualizado para teste",
		PaymentTerms:    "Net 45",
		ShippingAddress: "Rua Atualizada, 456 - São Paulo/SP",
		Items: []models.POItem{
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

	// Expect query to check if purchase order exists
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1 ORDER BY "purchase_orders"."id" LIMIT \$2`).
		WithArgs(orderID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "po_no", "so_no", "sales_order_id", "contact_id", "status",
			"created_at", "updated_at", "expected_date",
			"subtotal", "tax_total", "discount_total", "grand_total",
			"notes", "payment_terms", "shipping_address",
		}).AddRow(
			initialOrder.ID, initialOrder.PONo, initialOrder.SONo, initialOrder.SalesOrderID,
			initialOrder.ContactID, initialOrder.Status, initialOrder.CreatedAt, initialOrder.UpdatedAt,
			initialOrder.ExpectedDate, initialOrder.SubTotal, initialOrder.TaxTotal,
			initialOrder.DiscountTotal, initialOrder.GrandTotal, initialOrder.Notes,
			initialOrder.PaymentTerms, initialOrder.ShippingAddress,
		))

	// Expect purchase order update
	mock.ExpectExec(`UPDATE "purchase_orders" SET`).
		WithArgs(
			orderID,
			initialOrder.PONo,
			initialOrder.SONo,
			initialOrder.SalesOrderID,
			initialOrder.ContactID,
			models.POStatusConfirmed,
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
	mock.ExpectExec(`DELETE FROM "purchase_order_items" WHERE purchase_order_id = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect batch insert of new items
	mock.ExpectQuery(`INSERT INTO "purchase_order_items" \("purchase_order_id","product_id","product_name","product_code","description","quantity","unit_price","discount","tax","total"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10\),\(\$11,\$12,\$13,\$14,\$15,\$16,\$17,\$18,\$19,\$20\) RETURNING "id"`).
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
	err := repo.UpdatePurchaseOrder(orderID, updatedOrder)
	if err != nil {
		t.Fatalf("Error updating purchase order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	t.Run("Invalid ID", func(t *testing.T) {
		invalidID := -1
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1 ORDER BY "purchase_orders"."id" LIMIT \$2`).
			WithArgs(invalidID, 1).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		err = repo.UpdatePurchaseOrder(invalidID, updatedOrder)
		if err == nil {
			t.Errorf("Expected error for invalid ID, got nil")
		}
	})
}

// TestDeletePurchaseOrder tests deleting a purchase order by ID
func TestDeletePurchaseOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Test data
	orderID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect check for related deliveries
	mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries" WHERE purchase_order_id = \$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect deletion of items
	mock.ExpectExec(`DELETE FROM "purchase_order_items" WHERE purchase_order_id = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of purchase order
	mock.ExpectExec(`DELETE FROM "purchase_orders" WHERE "purchase_orders"."id" = \$1`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeletePurchaseOrder(orderID)
	if err != nil {
		t.Fatalf("Error deleting purchase order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with existing related records (deliveries)
	t.Run("With related records", func(t *testing.T) {
		orderID = 2
		mock.ExpectBegin()

		// Expect check for related deliveries - found 1 record
		mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries" WHERE purchase_order_id = \$1`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Expect rollback since related records exist
		mock.ExpectRollback()

		// Execute function to test
		err = repo.DeletePurchaseOrder(orderID)
		if err == nil {
			t.Errorf("Expected error for purchase order with related records, got nil")
		}

		// Verify error is ErrRelatedRecordsExist
		if fmt.Sprintf("%v", err) != fmt.Sprintf("%v: pedido possui 1 entregas vinculadas", errors.ErrRelatedRecordsExist) {
			t.Errorf("Expected error %v, got %v", errors.ErrRelatedRecordsExist, err)
		}
	})
}

// TestGetPurchaseOrdersByStatus tests retrieving purchase orders by status
func TestGetPurchaseOrdersByStatus(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Test data
	status := models.POStatusSent
	totalItems := int64(3)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "purchase_orders" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "po_no", "so_no", "sales_order_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test purchase orders to results
	for i := 1; i <= int(totalItems); i++ {
		testOrder := createTestPurchaseOrder(i)
		testOrder.Status = status
		rows.AddRow(
			testOrder.ID, testOrder.PONo, testOrder.SONo, testOrder.SalesOrderID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern with OrderBy expected_date ASC
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE status = \$1 ORDER BY expected_date ASC LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading associations for each purchase order
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "purchase_order_items" WHERE "purchase_order_items"."purchase_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "purchase_order_id", "product_id", "product_name", "product_code",
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

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "contact_id", "status",
			}).AddRow(
				1, "TEST-SO-001", 2, "confirmed",
			))
	}

	// Execute function to test
	result, err := repo.GetPurchaseOrdersByStatus(status, nil)
	if err != nil {
		t.Fatalf("Error getting purchase orders by status: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify purchase orders were returned with correct status
	orders, ok := result.Items.([]models.PurchaseOrder)
	if !ok {
		t.Fatalf("Could not convert items to []models.PurchaseOrder")
	}

	for i, o := range orders {
		if o.Status != status {
			t.Errorf("Purchase order %d has incorrect status. Expected: %s, Got: %s",
				i, status, o.Status)
		}
	}
}

// TestGetPurchaseOrdersBySalesOrder tests retrieving purchase orders by sales order ID
func TestGetPurchaseOrdersBySalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockPurchaseOrderRepo(db)

	// Test data
	salesOrderID := 1
	totalItems := 2

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "po_no", "so_no", "sales_order_id", "contact_id", "status",
		"created_at", "updated_at", "expected_date",
		"subtotal", "tax_total", "discount_total", "grand_total",
		"notes", "payment_terms", "shipping_address",
	})

	// Add test purchase orders to results
	for i := 1; i <= totalItems; i++ {
		testOrder := createTestPurchaseOrder(i)
		testOrder.SalesOrderID = salesOrderID
		rows.AddRow(
			testOrder.ID, testOrder.PONo, testOrder.SONo, testOrder.SalesOrderID,
			testOrder.ContactID, testOrder.Status, testOrder.CreatedAt, testOrder.UpdatedAt,
			testOrder.ExpectedDate, testOrder.SubTotal, testOrder.TaxTotal,
			testOrder.DiscountTotal, testOrder.GrandTotal, testOrder.Notes,
			testOrder.PaymentTerms, testOrder.ShippingAddress,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "purchase_orders" WHERE sales_order_id = \$1 ORDER BY created_at DESC`).
		WithArgs(salesOrderID).
		WillReturnRows(rows)

	// Expect queries for loading items and contacts for each purchase order
	for i := 1; i <= totalItems; i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "purchase_order_items" WHERE "purchase_order_items"."purchase_order_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "purchase_order_id", "product_id", "product_name", "product_code",
				"description", "quantity", "unit_price", "discount", "tax", "total",
			}).AddRow(
				1, i, 1, "Produto de Teste", "PROD-001",
				"Descrição do produto de teste", 10, 100.00, 5.00, 15.00, 1100.00,
			))

		// Contact query
		mock.ExpectQuery(`SELECT \* FROM "contacts" WHERE "contacts"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "email",
			}).AddRow(
				1, "Test Contact", "test@example.com",
			))
	}

	// Execute function to test
	orders, err := repo.GetPurchaseOrdersBySalesOrder(salesOrderID)
	if err != nil {
		t.Fatalf("Error getting purchase orders by sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify purchase orders count
	if len(orders) != totalItems {
		t.Errorf("Expected %d purchase orders, got %d", totalItems, len(orders))
	}

	// Verify purchase orders have correct sales order ID
	for i, o := range orders {
		if o.SalesOrderID != salesOrderID {
			t.Errorf("Purchase order %d has incorrect sales order ID. Expected: %d, Got: %d",
				i, salesOrderID, o.SalesOrderID)
		}
	}
}
