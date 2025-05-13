package dtos

import "time"

// DeliveryTrackingInfo representa informações detalhadas de rastreamento
type DeliveryTrackingInfo struct {
	DeliveryID      int                  `json:"delivery_id"`
	DeliveryNo      string               `json:"delivery_no"`
	Status          string               `json:"status"`
	TrackingNumber  string               `json:"tracking_number,omitempty"`
	ShippingMethod  string               `json:"shipping_method,omitempty"`
	ShippingAddress string               `json:"shipping_address"`
	DeliveryDate    time.Time            `json:"delivery_date"`
	ReceivedDate    *time.Time           `json:"received_date,omitempty"`
	Items           []DeliveryItemStatus `json:"items"`
	Events          []TrackingEvent      `json:"events"`
}

// DeliveryItemStatus representa o status de um item na entrega
type DeliveryItemStatus struct {
	ItemID      int    `json:"item_id"`
	ProductName string `json:"product_name"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	ReceivedQty int    `json:"received_qty"`
	PendingQty  int    `json:"pending_qty"`
	Status      string `json:"status"` // pending, partial, complete
}

// PaymentTrackingInfo representa informações de rastreamento de pagamento
type PaymentTrackingInfo struct {
	PaymentID     int             `json:"payment_id"`
	InvoiceID     int             `json:"invoice_id"`
	InvoiceNo     string          `json:"invoice_no"`
	Amount        float64         `json:"amount"`
	Status        string          `json:"status"`
	PaymentMethod string          `json:"payment_method"`
	Reference     string          `json:"reference,omitempty"`
	PaymentDate   time.Time       `json:"payment_date"`
	ReconcileDate *time.Time      `json:"reconcile_date,omitempty"`
	Events        []TrackingEvent `json:"events"`
}

// DocumentTrackingInfo representa informações de rastreamento de documento
type DocumentTrackingInfo struct {
	DocumentID    int                `json:"document_id"`
	DocumentType  string             `json:"document_type"`
	DocumentNo    string             `json:"document_no"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	StatusHistory []StatusTransition `json:"status_history"`
	RelatedDocs   []RelatedDocument  `json:"related_documents"`
	Activities    []ActivityLog      `json:"activities"`
}

// StatusTransition representa uma transição de status
type StatusTransition struct {
	ID         int       `json:"id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Timestamp  time.Time `json:"timestamp"`
	UserID     int       `json:"user_id,omitempty"`
	UserName   string    `json:"user_name,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

// TrackingEvent representa um evento de rastreamento
type TrackingEvent struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Location    string    `json:"location,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// RelatedDocument representa um documento relacionado
type RelatedDocument struct {
	DocumentID   int    `json:"document_id"`
	DocumentType string `json:"document_type"`
	DocumentNo   string `json:"document_no"`
	Status       string `json:"status"`
	Relationship string `json:"relationship"` // parent, child, reference
}

// ActivityLog representa um log de atividade
type ActivityLog struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	UserID      int       `json:"user_id,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Changes     []Change  `json:"changes,omitempty"`
}

// Change representa uma mudança em um campo
type Change struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// ShipmentTrackingDTO representa dados de rastreamento de envio
type ShipmentTrackingDTO struct {
	TrackingNumber    string          `json:"tracking_number"`
	Carrier           string          `json:"carrier"`
	Status            string          `json:"status"`
	LastUpdate        time.Time       `json:"last_update"`
	EstimatedDelivery *time.Time      `json:"estimated_delivery,omitempty"`
	CurrentLocation   string          `json:"current_location,omitempty"`
	Events            []ShipmentEvent `json:"events"`
}

// ShipmentEvent representa um evento de envio
type ShipmentEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}

// OrderTrackingDTO representa rastreamento completo de pedido
type OrderTrackingDTO struct {
	OrderID        int                 `json:"order_id"`
	OrderNo        string              `json:"order_no"`
	Status         string              `json:"status"`
	CustomerName   string              `json:"customer_name"`
	OrderDate      time.Time           `json:"order_date"`
	ExpectedDate   time.Time           `json:"expected_date"`
	TotalValue     float64             `json:"total_value"`
	Quotation      *DocumentStatusDTO  `json:"quotation,omitempty"`
	PurchaseOrders []DocumentStatusDTO `json:"purchase_orders,omitempty"`
	Deliveries     []DocumentStatusDTO `json:"deliveries,omitempty"`
	Invoices       []DocumentStatusDTO `json:"invoices,omitempty"`
	Payments       []PaymentStatusDTO  `json:"payments,omitempty"`
	Timeline       []TimelineEventDTO  `json:"timeline"`
}

// DocumentStatusDTO representa status resumido de documento
type DocumentStatusDTO struct {
	ID         int       `json:"id"`
	DocumentNo string    `json:"document_no"`
	Status     string    `json:"status"`
	Date       time.Time `json:"date"`
	Value      float64   `json:"value,omitempty"`
}

// PaymentStatusDTO representa status resumido de pagamento
type PaymentStatusDTO struct {
	ID            int       `json:"id"`
	Amount        float64   `json:"amount"`
	PaymentDate   time.Time `json:"payment_date"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
}

// TimelineEventDTO representa evento na timeline
type TimelineEventDTO struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
}
