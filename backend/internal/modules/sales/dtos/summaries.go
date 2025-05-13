package dtos

import "time"

// ContactSummaryBase contém campos base para resumos de contato
type ContactSummaryBase struct {
	ContactID      int       `json:"contact_id"`
	ContactName    string    `json:"contact_name"`
	ContactType    string    `json:"contact_type"`
	TotalCount     int       `json:"total_count"`
	TotalValue     float64   `json:"total_value"`
	LastUpdateDate time.Time `json:"last_update_date"`
}

// ContactDeliveriesSummary representa um resumo das deliveries de um contato
type ContactDeliveriesSummary struct {
	ContactSummaryBase
	DeliveryType   string  `json:"delivery_type"`
	PendingCount   int     `json:"pending_count"`
	ShippedCount   int     `json:"shipped_count"`
	DeliveredCount int     `json:"delivered_count"`
	ReturnedCount  int     `json:"returned_count"`
	OverdueCount   int     `json:"overdue_count"`
	DeliveryRate   float64 `json:"delivery_rate"`
	ReturnRate     float64 `json:"return_rate"`
}

// ContactInvoicesSummary representa um resumo das invoices de um contato
type ContactInvoicesSummary struct {
	ContactSummaryBase
	TotalPaid      float64 `json:"total_paid"`
	TotalPending   float64 `json:"total_pending"`
	OverdueCount   int     `json:"overdue_count"`
	OverdueValue   float64 `json:"overdue_value"`
	PaymentRate    float64 `json:"payment_rate"`
	AveragePayTime float64 `json:"average_payment_time_days"`
}

// ContactQuotationsSummary representa um resumo das quotations de um contato
type ContactQuotationsSummary struct {
	ContactSummaryBase
	TotalAccepted  float64 `json:"total_accepted"`
	TotalRejected  float64 `json:"total_rejected"`
	PendingCount   int     `json:"pending_count"`
	PendingValue   float64 `json:"pending_value"`
	ConversionRate float64 `json:"conversion_rate"`
	ExpiringCount  int     `json:"expiring_count"`
}

// ContactSalesOrdersSummary representa um resumo dos sales orders de um contato
type ContactSalesOrdersSummary struct {
	ContactSummaryBase
	TotalCompleted  float64 `json:"total_completed"`
	TotalCancelled  float64 `json:"total_cancelled"`
	PendingCount    int     `json:"pending_count"`
	PendingValue    float64 `json:"pending_value"`
	FulfillmentRate float64 `json:"fulfillment_rate"`
	AverageValue    float64 `json:"average_value"`
}

// ContactPurchaseOrdersSummary representa um resumo dos purchase orders de um contato
type ContactPurchaseOrdersSummary struct {
	ContactSummaryBase
	TotalReceived   float64 `json:"total_received"`
	TotalCancelled  float64 `json:"total_cancelled"`
	PendingCount    int     `json:"pending_count"`
	PendingValue    float64 `json:"pending_value"`
	OverdueCount    int     `json:"overdue_count"`
	OverdueValue    float64 `json:"overdue_value"`
	FulfillmentRate float64 `json:"fulfillment_rate"`
}

// ContactSalesProcessSummary representa um resumo dos processos de um contato
type ContactSalesProcessSummary struct {
	ContactSummaryBase
	ActiveProcesses    int     `json:"active_processes"`
	CompletedProcesses int     `json:"completed_processes"`
	TotalProfit        float64 `json:"total_profit"`
	AverageValue       float64 `json:"average_value"`
	ConversionRate     float64 `json:"conversion_rate"`
	AverageCycleTime   float64 `json:"average_cycle_time_days"`
}

// ContactOverallSummary representa um resumo completo de um contato
type ContactOverallSummary struct {
	ContactInfo      ContactBasicInfo             `json:"contact_info"`
	Quotations       ContactQuotationsSummary     `json:"quotations"`
	SalesOrders      ContactSalesOrdersSummary    `json:"sales_orders"`
	PurchaseOrders   ContactPurchaseOrdersSummary `json:"purchase_orders"`
	Deliveries       ContactDeliveriesSummary     `json:"deliveries"`
	Invoices         ContactInvoicesSummary       `json:"invoices"`
	SalesProcesses   ContactSalesProcessSummary   `json:"sales_processes"`
	TotalRevenue     float64                      `json:"total_revenue"`
	TotalProfit      float64                      `json:"total_profit"`
	ProfitMargin     float64                      `json:"profit_margin_percentage"`
	CustomerLifetime float64                      `json:"customer_lifetime_value"`
	LastActivity     time.Time                    `json:"last_activity"`
}

// PeriodSummary representa um resumo por período
type PeriodSummary struct {
	Period      string  `json:"period"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Count       int     `json:"count"`
	TotalValue  float64 `json:"total_value"`
	TotalProfit float64 `json:"total_profit,omitempty"`
	Growth      float64 `json:"growth_percentage,omitempty"`
}

// DashboardSummary representa um resumo para dashboard
type DashboardSummary struct {
	Period           string               `json:"period"`
	Revenue          AmountSummary        `json:"revenue"`
	Profit           AmountSummary        `json:"profit"`
	Conversions      ConversionMetrics    `json:"conversions"`
	TopCustomers     []CustomerSummary    `json:"top_customers"`
	TopProducts      []ProductSummary     `json:"top_products"`
	PendingTasks     []PendingTaskSummary `json:"pending_tasks"`
	RecentActivities []ActivitySummary    `json:"recent_activities"`
}

// CustomerSummary representa um resumo de cliente
type CustomerSummary struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Revenue  float64 `json:"revenue"`
	Profit   float64 `json:"profit"`
	Orders   int     `json:"orders"`
	LastDate string  `json:"last_date"`
}

// ProductSummary representa um resumo de produto
type ProductSummary struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	Revenue  float64 `json:"revenue"`
	Quantity int     `json:"quantity"`
	Orders   int     `json:"orders"`
}

// PendingTaskSummary representa um resumo de tarefa pendente
type PendingTaskSummary struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	IsOverdue   bool      `json:"is_overdue"`
	Priority    string    `json:"priority"`
}

// ActivitySummary representa um resumo de atividade
type ActivitySummary struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	UserName    string    `json:"user_name,omitempty"`
	DocumentNo  string    `json:"document_no,omitempty"`
}
