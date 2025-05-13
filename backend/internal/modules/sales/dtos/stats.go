package dtos

// DeliveryStats representa estatísticas de deliveries
type DeliveryStats struct {
	TotalDeliveries     int                `json:"total_deliveries"`
	TotalPending        int                `json:"total_pending"`
	TotalShipped        int                `json:"total_shipped"`
	TotalDelivered      int                `json:"total_delivered"`
	TotalReturned       int                `json:"total_returned"`
	CountByStatus       map[string]int     `json:"count_by_status"`
	DeliveryRate        float64            `json:"delivery_rate"`
	ReturnRate          float64            `json:"return_rate"`
	AverageDeliveryTime float64            `json:"average_delivery_time_days"`
	ValueByStatus       map[string]float64 `json:"value_by_status,omitempty"`
}

// InvoiceStats representa estatísticas de invoices
type InvoiceStats struct {
	TotalInvoices int                `json:"total_invoices"`
	TotalValue    float64            `json:"total_value"`
	TotalPaid     float64            `json:"total_paid"`
	TotalPending  float64            `json:"total_pending"`
	TotalOverdue  float64            `json:"total_overdue"`
	CountByStatus map[string]int     `json:"count_by_status"`
	ValueByStatus map[string]float64 `json:"value_by_status"`
	PaymentRate   float64            `json:"payment_rate"`
}

// PaymentStats representa estatísticas de payments
type PaymentStats struct {
	TotalPayments     int                `json:"total_payments"`
	TotalAmount       float64            `json:"total_amount"`
	AverageAmount     float64            `json:"average_amount"`
	CountByMethod     map[string]int     `json:"count_by_method"`
	AmountByMethod    map[string]float64 `json:"amount_by_method"`
	TodayPayments     int                `json:"today_payments"`
	TodayAmount       float64            `json:"today_amount"`
	ThisMonthPayments int                `json:"this_month_payments"`
	ThisMonthAmount   float64            `json:"this_month_amount"`
}

// PurchaseOrderStats representa estatísticas de purchase orders
type PurchaseOrderStats struct {
	TotalOrders     int                `json:"total_orders"`
	TotalValue      float64            `json:"total_value"`
	TotalDraft      float64            `json:"total_draft"`
	TotalSent       float64            `json:"total_sent"`
	TotalConfirmed  float64            `json:"total_confirmed"`
	TotalReceived   float64            `json:"total_received"`
	TotalCancelled  float64            `json:"total_cancelled"`
	CountByStatus   map[string]int     `json:"count_by_status"`
	ValueByStatus   map[string]float64 `json:"value_by_status"`
	FulfillmentRate float64            `json:"fulfillment_rate"`
	AverageValue    float64            `json:"average_value"`
}

// QuotationStats representa estatísticas de quotations
type QuotationStats struct {
	TotalQuotations int                `json:"total_quotations"`
	TotalValue      float64            `json:"total_value"`
	TotalAccepted   float64            `json:"total_accepted"`
	TotalRejected   float64            `json:"total_rejected"`
	TotalPending    float64            `json:"total_pending"`
	TotalExpired    float64            `json:"total_expired"`
	CountByStatus   map[string]int     `json:"count_by_status"`
	ValueByStatus   map[string]float64 `json:"value_by_status"`
	ConversionRate  float64            `json:"conversion_rate"`
	AverageValue    float64            `json:"average_value"`
}

// SalesOrderStats representa estatísticas de sales orders
type SalesOrderStats struct {
	TotalOrders     int                `json:"total_orders"`
	TotalValue      float64            `json:"total_value"`
	TotalConfirmed  float64            `json:"total_confirmed"`
	TotalProcessing float64            `json:"total_processing"`
	TotalCompleted  float64            `json:"total_completed"`
	TotalCancelled  float64            `json:"total_cancelled"`
	CountByStatus   map[string]int     `json:"count_by_status"`
	ValueByStatus   map[string]float64 `json:"value_by_status"`
	FulfillmentRate float64            `json:"fulfillment_rate"`
	AverageValue    float64            `json:"average_value"`
}

// SalesProcessStats representa estatísticas de sales processes
type SalesProcessStats struct {
	TotalProcesses   int                `json:"total_processes"`
	TotalValue       float64            `json:"total_value"`
	TotalProfit      float64            `json:"total_profit"`
	AverageValue     float64            `json:"average_value"`
	AverageProfit    float64            `json:"average_profit"`
	ProfitMargin     float64            `json:"profit_margin_percentage"`
	CountByStatus    map[string]int     `json:"count_by_status"`
	ValueByStatus    map[string]float64 `json:"value_by_status"`
	CompletionRate   float64            `json:"completion_rate"`
	AverageCycleTime float64            `json:"average_cycle_time_days"`
}

// PaymentMethodStats representa estatísticas por método de pagamento
type PaymentMethodStats struct {
	Method        string  `json:"method"`
	Count         int     `json:"count"`
	TotalAmount   float64 `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
	Percentage    float64 `json:"percentage"`
}

// DailyStats representa estatísticas diárias genéricas
type DailyStats struct {
	Date       string  `json:"date"`
	Count      int     `json:"count"`
	TotalValue float64 `json:"total_value"`
}

// MonthlyStats representa estatísticas mensais genéricas
type MonthlyStats struct {
	Year       int           `json:"year"`
	Month      int           `json:"month"`
	Count      int           `json:"count"`
	TotalValue float64       `json:"total_value"`
	ByDay      []DailyStats  `json:"by_day,omitempty"`
	Growth     GrowthMetrics `json:"growth,omitempty"`
}

// GrowthMetrics representa métricas de crescimento
type GrowthMetrics struct {
	CountGrowth   float64 `json:"count_growth_percentage"`
	ValueGrowth   float64 `json:"value_growth_percentage"`
	PreviousCount int     `json:"previous_count"`
	PreviousValue float64 `json:"previous_value"`
}

// ConversionMetrics representa métricas de conversão
type ConversionMetrics struct {
	TotalQuotations       int                     `json:"total_quotations"`
	QuotationToSORate     float64                 `json:"quotation_to_so_rate"`
	SOToInvoiceRate       float64                 `json:"so_to_invoice_rate"`
	InvoiceToPaymentRate  float64                 `json:"invoice_to_payment_rate"`
	OverallConversionRate float64                 `json:"overall_conversion_rate"`
	AverageConversionTime float64                 `json:"average_conversion_time_days"`
	ByStage               map[string]StageMetrics `json:"by_stage"`
}

// StageMetrics representa métricas por estágio
type StageMetrics struct {
	Count           int     `json:"count"`
	ConversionRate  float64 `json:"conversion_rate"`
	AverageTime     float64 `json:"average_time_days"`
	AbandonmentRate float64 `json:"abandonment_rate"`
}
