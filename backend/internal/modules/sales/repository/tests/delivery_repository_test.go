package repository

import (
	db_config "ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
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

// createMockDeliveryRepo creates a repository with mocked DB
func createMockDeliveryRepo(db *gorm.DB) repository.DeliveryRepository {
	return repository.NewTestDeliveryRepository(db)
}

// createTestDelivery creates a test delivery with given ID
func createTestDelivery(id int) *models.Delivery {
	return &models.Delivery{
		ID:              id,
		DeliveryNo:      fmt.Sprintf("TEST-DEL-%03d", id),
		SalesOrderID:    1,
		SONo:            "SO-001",
		PurchaseOrderID: 0, // No purchase order for this test
		Status:          models.DeliveryStatusPending,
		DeliveryDate:    time.Now().AddDate(0, 0, 5),
		ShippingMethod:  "Standard Shipping",
		TrackingNumber:  "TRK12345",
		ShippingAddress: "123 Test Street, Test City",
		Notes:           "Test delivery notes",
		Items: []models.DeliveryItem{
			{
				ID:          1,
				DeliveryID:  id,
				ProductID:   1,
				ProductName: "Test Product",
				ProductCode: "PROD-001",
				Description: "Test product description",
				Quantity:    5,
				ReceivedQty: 0,
				Notes:       "Item notes",
			},
		},
		SalesOrder: &models.SalesOrder{
			ID:     1,
			SONo:   "SO-001",
			Status: models.SOStatusConfirmed,
		},
	}
}

// TestCreateDelivery tests the creation of a new delivery
func TestCreateDelivery(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Create test delivery
	delivery := createTestDelivery(0) // ID 0 because it will be assigned by the DB

	// Remove relationships to avoid GORM trying to save them
	delivery.SalesOrder = nil

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect delivery insert
	mock.ExpectQuery(`INSERT INTO "deliveries"`).
		WithArgs(
			sqlmock.AnyArg(), // delivery_no
			sqlmock.AnyArg(), // purchase_order_id
			sqlmock.AnyArg(), // po_no
			sqlmock.AnyArg(), // sales_order_id
			sqlmock.AnyArg(), // so_no
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // delivery_date
			sqlmock.AnyArg(), // received_date
			sqlmock.AnyArg(), // shipping_method
			sqlmock.AnyArg(), // tracking_number
			sqlmock.AnyArg(), // shipping_address
			sqlmock.AnyArg(), // notes
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect item insert
	for range delivery.Items {
		mock.ExpectQuery(`INSERT INTO "delivery_items"`).
			WithArgs(
				sqlmock.AnyArg(), // delivery_id
				sqlmock.AnyArg(), // product_id
				sqlmock.AnyArg(), // product_name
				sqlmock.AnyArg(), // product_code
				sqlmock.AnyArg(), // description
				sqlmock.AnyArg(), // quantity
				sqlmock.AnyArg(), // received_qty
				sqlmock.AnyArg(), // notes
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	// Expect query to retrieve items
	mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE delivery_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "delivery_id", "product_id", "product_name", "product_code",
			"description", "quantity", "received_qty", "notes",
		}).AddRow(
			1, 1, 1, "Test Product", "PROD-001",
			"Test product description", 5, 0, "Item notes",
		))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	if err := repo.CreateDelivery(delivery); err != nil {
		t.Fatalf("Error creating delivery: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify delivery ID was set
	if delivery.ID != 1 {
		t.Errorf("Expected delivery ID to be 1, got %d", delivery.ID)
	}

	// Verify items were loaded
	if len(delivery.Items) == 0 {
		t.Fatalf("Delivery items were not loaded")
	}
}

// TestGetDeliveryByID tests retrieving a delivery by ID
func TestGetDeliveryByID(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Setup test data
	testDelivery := createTestDelivery(1)
	deliveryID := testDelivery.ID

	// Setup expectations for finding delivery
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE "deliveries"."id" = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
			"status", "created_at", "updated_at", "delivery_date", "received_date",
			"shipping_method", "tracking_number", "shipping_address", "notes",
		}).AddRow(
			testDelivery.ID, testDelivery.DeliveryNo, testDelivery.PurchaseOrderID,
			testDelivery.PONo, testDelivery.SalesOrderID, testDelivery.SONo,
			testDelivery.Status, time.Now(), time.Now(), testDelivery.DeliveryDate,
			testDelivery.ReceivedDate, testDelivery.ShippingMethod, testDelivery.TrackingNumber,
			testDelivery.ShippingAddress, testDelivery.Notes,
		))

	// Setup expectations for loading items
	mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
		WithArgs(deliveryID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "delivery_id", "product_id", "product_name", "product_code",
			"description", "quantity", "received_qty", "notes",
		}).AddRow(
			testDelivery.Items[0].ID, testDelivery.Items[0].DeliveryID,
			testDelivery.Items[0].ProductID, testDelivery.Items[0].ProductName,
			testDelivery.Items[0].ProductCode, testDelivery.Items[0].Description,
			testDelivery.Items[0].Quantity, testDelivery.Items[0].ReceivedQty,
			testDelivery.Items[0].Notes,
		))

	// Setup expectations for loading sales order
	mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
		WithArgs(testDelivery.SalesOrderID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "so_no", "status",
		}).AddRow(
			testDelivery.SalesOrder.ID,
			testDelivery.SalesOrder.SONo,
			testDelivery.SalesOrder.Status,
		))

	// Execute function to test
	retrievedDelivery, err := repo.GetDeliveryByID(deliveryID)
	if err != nil {
		t.Fatalf("Error retrieving delivery: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify retrieved delivery data
	if retrievedDelivery.ID != deliveryID {
		t.Errorf("Expected delivery ID %d, got %d", deliveryID, retrievedDelivery.ID)
	}

	if retrievedDelivery.DeliveryNo != testDelivery.DeliveryNo {
		t.Errorf("Expected delivery number %s, got %s",
			testDelivery.DeliveryNo, retrievedDelivery.DeliveryNo)
	}

	// Verify items were loaded
	if len(retrievedDelivery.Items) == 0 {
		t.Errorf("Items were not loaded")
	}

	// Verify sales order was loaded
	if retrievedDelivery.SalesOrder == nil {
		t.Errorf("Sales order was not loaded")
	}

	// Test invalid ID case
	invalidID := -1
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE "deliveries"."id" = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = repo.GetDeliveryByID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetAllDeliveries tests retrieving all deliveries
func TestGetAllDeliveries(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	totalItems := int64(3)
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
		"status", "created_at", "updated_at", "delivery_date", "received_date",
		"shipping_method", "tracking_number", "shipping_address", "notes",
	})

	// Add 3 test deliveries to results
	for i := 1; i <= int(totalItems); i++ {
		testDel := createTestDelivery(i)
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "deliveries" LIMIT \$1`).
		WithArgs(pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items and related orders for each delivery
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, i, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "status",
			}).AddRow(
				1, "SO-001", models.SOStatusConfirmed,
			))
	}

	// Execute function to test
	result, err := repo.GetAllDeliveries(nil) // nil for default pagination
	if err != nil {
		t.Fatalf("Error getting all deliveries: %v", err)
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

	// Verify deliveries were returned
	deliveries, ok := result.Items.([]models.Delivery)
	if !ok {
		t.Fatalf("Could not convert items to []models.Delivery")
	}

	if len(deliveries) != int(totalItems) {
		t.Errorf("Expected %d deliveries, got %d", totalItems, len(deliveries))
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
		mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

		// Setup expectations for retrieval query with custom pagination
		rows := sqlmock.NewRows([]string{
			"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
			"status", "created_at", "updated_at", "delivery_date", "received_date",
			"shipping_method", "tracking_number", "shipping_address", "notes",
		})

		// Add just 1 delivery to results (page 2, size 1)
		testDel := createTestDelivery(2) // Second item for page 2
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)

		mock.ExpectQuery(`SELECT \* FROM "deliveries" LIMIT \$1 OFFSET \$2`).
			WithArgs(customPageSize, expectedOffset).
			WillReturnRows(rows)

		// Expect queries for loading items and related orders
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, 2, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "status",
			}).AddRow(
				1, "SO-001", models.SOStatusConfirmed,
			))

		// Execute function to test with custom pagination
		result, err := repo.GetAllDeliveries(params)
		if err != nil {
			t.Fatalf("Error getting deliveries with custom pagination: %v", err)
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

		// Verify deliveries count
		deliveries, ok := result.Items.([]models.Delivery)
		if !ok {
			t.Fatalf("Could not convert items to []models.Delivery")
		}

		if len(deliveries) != customPageSize {
			t.Errorf("Expected %d deliveries, got %d", customPageSize, len(deliveries))
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
		result, err := repo.GetAllDeliveries(invalidParams)

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

// TestUpdateDelivery tests updating an existing delivery
func TestUpdateDelivery(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	deliveryID := 1
	initialDelivery := createTestDelivery(deliveryID)

	// Create updated delivery data
	updatedDelivery := &models.Delivery{
		ID:              deliveryID,
		DeliveryNo:      initialDelivery.DeliveryNo,
		SalesOrderID:    initialDelivery.SalesOrderID,
		SONo:            initialDelivery.SONo,
		Status:          models.DeliveryStatusShipped, // Updated status
		DeliveryDate:    time.Now().AddDate(0, 0, 3),  // Updated delivery date
		ShippingMethod:  "Express Shipping",           // Updated shipping method
		TrackingNumber:  "EXPTRK6789",                 // Updated tracking number
		ShippingAddress: "456 Updated Street, Test City",
		Notes:           "Updated delivery notes",
		Items: []models.DeliveryItem{
			{
				ProductID:   1,
				ProductName: "Updated Product",
				ProductCode: "PROD-001",
				Description: "Updated product description",
				Quantity:    8, // Updated quantity
				ReceivedQty: 5, // Updated received quantity
				Notes:       "Updated item notes",
			},
			{
				// New item
				ProductID:   2,
				ProductName: "Additional Product",
				ProductCode: "PROD-002",
				Description: "Additional product description",
				Quantity:    3,
				ReceivedQty: 0,
				Notes:       "Additional item notes",
			},
		},
	}

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect query to check if delivery exists
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE "deliveries"."id" = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(deliveryID, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
			"status", "created_at", "updated_at", "delivery_date", "received_date",
			"shipping_method", "tracking_number", "shipping_address", "notes",
		}).AddRow(
			initialDelivery.ID, initialDelivery.DeliveryNo, initialDelivery.PurchaseOrderID,
			initialDelivery.PONo, initialDelivery.SalesOrderID, initialDelivery.SONo,
			initialDelivery.Status, time.Now(), time.Now(), initialDelivery.DeliveryDate,
			initialDelivery.ReceivedDate, initialDelivery.ShippingMethod, initialDelivery.TrackingNumber,
			initialDelivery.ShippingAddress, initialDelivery.Notes,
		))

	// Expect delivery update - match exactly with the 12 parameters in the actual query
	mock.ExpectExec(`UPDATE "deliveries" SET "id"=\$1,"delivery_no"=\$2,"sales_order_id"=\$3,"so_no"=\$4,"status"=\$5,"updated_at"=\$6,"delivery_date"=\$7,"shipping_method"=\$8,"tracking_number"=\$9,"shipping_address"=\$10,"notes"=\$11 WHERE "id" = \$12`).
		WithArgs(
			deliveryID,
			initialDelivery.DeliveryNo,
			initialDelivery.SalesOrderID,
			initialDelivery.SONo,
			models.DeliveryStatusShipped,
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // delivery_date
			"Express Shipping",
			"EXPTRK6789",
			"456 Updated Street, Test City",
			"Updated delivery notes",
			deliveryID, // WHERE id = ?
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect deletion of existing items
	mock.ExpectExec(`DELETE FROM "delivery_items" WHERE delivery_id = \$1`).
		WithArgs(deliveryID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect batch insertion of new items
	mock.ExpectQuery(`INSERT INTO "delivery_items" \("delivery_id","product_id","product_name","product_code","description","quantity","received_qty","notes"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\),\(\$9,\$10,\$11,\$12,\$13,\$14,\$15,\$16\) RETURNING "id"`).
		WithArgs(
			// First item
			deliveryID,
			updatedDelivery.Items[0].ProductID,
			updatedDelivery.Items[0].ProductName,
			updatedDelivery.Items[0].ProductCode,
			updatedDelivery.Items[0].Description,
			updatedDelivery.Items[0].Quantity,
			updatedDelivery.Items[0].ReceivedQty,
			updatedDelivery.Items[0].Notes,
			// Second item
			deliveryID,
			updatedDelivery.Items[1].ProductID,
			updatedDelivery.Items[1].ProductName,
			updatedDelivery.Items[1].ProductCode,
			updatedDelivery.Items[1].Description,
			updatedDelivery.Items[1].Quantity,
			updatedDelivery.Items[1].ReceivedQty,
			updatedDelivery.Items[1].Notes,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.UpdateDelivery(deliveryID, updatedDelivery)
	if err != nil {
		t.Fatalf("Error updating delivery: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with invalid ID
	invalidID := -1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE "deliveries"."id" = \$1 ORDER BY "deliveries"."id" LIMIT \$2`).
		WithArgs(invalidID, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err = repo.UpdateDelivery(invalidID, updatedDelivery)
	if err == nil {
		t.Errorf("Expected error for invalid ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestDeleteDelivery tests deleting a delivery by ID
func TestDeleteDelivery(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	deliveryID := 1

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect deletion of items
	mock.ExpectExec(`DELETE FROM "delivery_items" WHERE delivery_id = \$1`).
		WithArgs(deliveryID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of delivery
	mock.ExpectExec(`DELETE FROM "deliveries" WHERE "deliveries"."id" = \$1`).
		WithArgs(deliveryID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	// Execute function to test
	err := repo.DeleteDelivery(deliveryID)
	if err != nil {
		t.Fatalf("Error deleting delivery: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Test with non-existent ID
	nonExistentID := 999
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "delivery_items" WHERE delivery_id = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "deliveries" WHERE "deliveries"."id" = \$1`).
		WithArgs(nonExistentID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.DeleteDelivery(nonExistentID)
	if err == nil {
		t.Errorf("Expected error for non-existent ID, got nil")
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetDeliveriesByStatus tests retrieving deliveries by status with pagination
func TestGetDeliveriesByStatus(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	status := models.DeliveryStatusPending
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
		"status", "created_at", "updated_at", "delivery_date", "received_date",
		"shipping_method", "tracking_number", "shipping_address", "notes",
	})

	// Add test deliveries to results
	for i := 1; i <= int(totalItems); i++ {
		testDel := createTestDelivery(i)
		testDel.Status = status
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE status = \$1 LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items and related orders
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, i, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "status",
			}).AddRow(
				1, "SO-001", models.SOStatusConfirmed,
			))
	}

	// Execute function to test
	result, err := repo.GetDeliveriesByStatus(status, nil)
	if err != nil {
		t.Fatalf("Error getting deliveries by status: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify deliveries were returned with correct status
	deliveries, ok := result.Items.([]models.Delivery)
	if !ok {
		t.Fatalf("Could not convert items to []models.Delivery")
	}

	for i, d := range deliveries {
		if d.Status != status {
			t.Errorf("Delivery %d has incorrect status. Expected: %s, Got: %s",
				i, status, d.Status)
		}
	}
}

// TestGetDeliveriesBySalesOrder tests retrieving deliveries by sales order ID
func TestGetDeliveriesBySalesOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	salesOrderID := 1
	totalItems := 2

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
		"status", "created_at", "updated_at", "delivery_date", "received_date",
		"shipping_method", "tracking_number", "shipping_address", "notes",
	})

	// Add test deliveries to results
	for i := 1; i <= totalItems; i++ {
		testDel := createTestDelivery(i)
		testDel.SalesOrderID = salesOrderID
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE sales_order_id = \$1`).
		WithArgs(salesOrderID).
		WillReturnRows(rows)

	// Expect queries for loading items for each delivery
	for i := 1; i <= totalItems; i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, i, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))
	}

	// Execute function to test
	deliveries, err := repo.GetDeliveriesBySalesOrder(salesOrderID)
	if err != nil {
		t.Fatalf("Error getting deliveries by sales order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify deliveries were returned with correct sales order ID
	if len(deliveries) != totalItems {
		t.Errorf("Expected %d deliveries, got %d", totalItems, len(deliveries))
	}

	for i, d := range deliveries {
		if d.SalesOrderID != salesOrderID {
			t.Errorf("Delivery %d has incorrect sales order ID. Expected: %d, Got: %d",
				i, salesOrderID, d.SalesOrderID)
		}
	}
}

// TestGetDeliveriesByPurchaseOrder tests retrieving deliveries by purchase order ID
func TestGetDeliveriesByPurchaseOrder(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	purchaseOrderID := 5
	totalItems := 2

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
		"status", "created_at", "updated_at", "delivery_date", "received_date",
		"shipping_method", "tracking_number", "shipping_address", "notes",
	})

	// Add test deliveries to results
	for i := 1; i <= totalItems; i++ {
		testDel := createTestDelivery(i)
		testDel.PurchaseOrderID = purchaseOrderID
		testDel.PONo = fmt.Sprintf("PO-%03d", purchaseOrderID)
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE purchase_order_id = \$1`).
		WithArgs(purchaseOrderID).
		WillReturnRows(rows)

	// Expect queries for loading items for each delivery
	for i := 1; i <= totalItems; i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, i, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))
	}

	// Execute function to test
	deliveries, err := repo.GetDeliveriesByPurchaseOrder(purchaseOrderID)
	if err != nil {
		t.Fatalf("Error getting deliveries by purchase order: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify deliveries were returned with correct purchase order ID
	if len(deliveries) != totalItems {
		t.Errorf("Expected %d deliveries, got %d", totalItems, len(deliveries))
	}

	for i, d := range deliveries {
		if d.PurchaseOrderID != purchaseOrderID {
			t.Errorf("Delivery %d has incorrect purchase order ID. Expected: %d, Got: %d",
				i, purchaseOrderID, d.PurchaseOrderID)
		}
	}
}

// TestGetPendingDeliveries tests retrieving pending deliveries
func TestGetPendingDeliveries(t *testing.T) {
	// Setup mock database
	db, mock, sqlDB := db_config.SetupMockDB(t)
	defer sqlDB.Close()

	// Create repository with mock
	repo := createMockDeliveryRepo(db)

	// Test data
	status := models.DeliveryStatusPending
	totalItems := int64(2)
	pageSize := pagination.DefaultPageSize

	// Setup expectations for count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "deliveries" WHERE status = \$1`).
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalItems))

	// Setup expectations for retrieval query
	rows := sqlmock.NewRows([]string{
		"id", "delivery_no", "purchase_order_id", "po_no", "sales_order_id", "so_no",
		"status", "created_at", "updated_at", "delivery_date", "received_date",
		"shipping_method", "tracking_number", "shipping_address", "notes",
	})

	// Add test deliveries to results
	for i := 1; i <= int(totalItems); i++ {
		testDel := createTestDelivery(i)
		testDel.Status = status
		rows.AddRow(
			testDel.ID, testDel.DeliveryNo, testDel.PurchaseOrderID,
			testDel.PONo, testDel.SalesOrderID, testDel.SONo,
			testDel.Status, time.Now(), time.Now(), testDel.DeliveryDate,
			testDel.ReceivedDate, testDel.ShippingMethod, testDel.TrackingNumber,
			testDel.ShippingAddress, testDel.Notes,
		)
	}

	// Match the actual SQL query pattern
	mock.ExpectQuery(`SELECT \* FROM "deliveries" WHERE status = \$1 LIMIT \$2`).
		WithArgs(status, pageSize).
		WillReturnRows(rows)

	// Expect queries for loading items and related orders
	for i := 1; i <= int(totalItems); i++ {
		// Items query
		mock.ExpectQuery(`SELECT \* FROM "delivery_items" WHERE "delivery_items"."delivery_id" = \$1`).
			WithArgs(i).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "delivery_id", "product_id", "product_name", "product_code",
				"description", "quantity", "received_qty", "notes",
			}).AddRow(
				1, i, 1, "Test Product", "PROD-001",
				"Test product description", 5, 0, "Item notes",
			))

		// Sales order query
		mock.ExpectQuery(`SELECT \* FROM "sales_orders" WHERE "sales_orders"."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "so_no", "status",
			}).AddRow(
				1, "SO-001", models.SOStatusConfirmed,
			))
	}

	// Execute function to test
	result, err := repo.GetPendingDeliveries(nil)
	if err != nil {
		t.Fatalf("Error getting pending deliveries: %v", err)
	}

	// Verify expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}

	// Verify deliveries were returned with correct status
	deliveries, ok := result.Items.([]models.Delivery)
	if !ok {
		t.Fatalf("Could not convert items to []models.Delivery")
	}

	for i, d := range deliveries {
		if d.Status != status {
			t.Errorf("Delivery %d has incorrect status. Expected: %s, Got: %s",
				i, status, d.Status)
		}
	}
}
