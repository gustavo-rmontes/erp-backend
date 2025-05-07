package models

// Define status constants for each stage
const (
	// Quotation statuses
	QuotationStatusDraft     = "draft"
	QuotationStatusSent      = "sent"
	QuotationStatusAccepted  = "accepted"
	QuotationStatusRejected  = "rejected"
	QuotationStatusExpired   = "expired"
	QuotationStatusCancelled = "cancelled"

	// Sales Order statuses
	SOStatusDraft      = "draft"
	SOStatusConfirmed  = "confirmed"
	SOStatusProcessing = "processing"
	SOStatusCompleted  = "completed"
	SOStatusCancelled  = "cancelled"

	// Purchase Order statuses
	POStatusDraft     = "draft"
	POStatusSent      = "sent"
	POStatusConfirmed = "confirmed"
	POStatusReceived  = "received"
	POStatusCancelled = "cancelled"

	// Delivery statuses
	DeliveryStatusPending   = "pending"
	DeliveryStatusShipped   = "shipped"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusReturned  = "returned"

	// Invoice statuses
	InvoiceStatusDraft     = "draft"
	InvoiceStatusSent      = "sent"
	InvoiceStatusPartial   = "partial"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusOverdue   = "overdue"
	InvoiceStatusCancelled = "cancelled"
)
