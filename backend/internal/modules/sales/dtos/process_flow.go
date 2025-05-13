package dtos

import "time"

// CompleteProcessFlow representa o fluxo completo de um processo
type CompleteProcessFlow struct {
	Process        SalesProcessResponseDTO    `json:"process"`
	Quotation      *QuotationResponseDTO      `json:"quotation,omitempty"`
	SalesOrder     *SalesOrderResponseDTO     `json:"sales_order,omitempty"`
	PurchaseOrders []PurchaseOrderResponseDTO `json:"purchase_orders,omitempty"`
	Deliveries     []DeliveryResponseDTO      `json:"deliveries,omitempty"`
	Invoices       []InvoiceResponseDTO       `json:"invoices,omitempty"`
	Payments       []PaymentResponseDTO       `json:"payments,omitempty"`
	Timeline       []ProcessEvent             `json:"timeline"`
	Relationships  ProcessRelationships       `json:"relationships"`
}

// ProcessTimeline representa a linha do tempo de eventos do processo
type ProcessTimeline struct {
	ProcessID  int            `json:"process_id"`
	Events     []ProcessEvent `json:"events"`
	Duration   int            `json:"duration_days"`
	Status     string         `json:"current_status"`
	Milestones []Milestone    `json:"milestones"`
}

// ProcessEvent representa um evento na linha do tempo
type ProcessEvent struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	DocumentID  int       `json:"document_id,omitempty"`
	DocumentNo  string    `json:"document_no,omitempty"`
	Value       float64   `json:"value,omitempty"`
	UserName    string    `json:"user_name,omitempty"`
	Status      string    `json:"status,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
}

// Milestone representa um marco no processo
type Milestone struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AchievedAt  time.Time `json:"achieved_at,omitempty"`
	Status      string    `json:"status"`
	Order       int       `json:"order"`
}

// ProcessStage representa um estágio do processo
type ProcessStage struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	Duration    int       `json:"duration_days,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsCompleted bool      `json:"is_completed"`
	Progress    float64   `json:"progress_percentage"`
}

// ProcessRelationships representa os relacionamentos entre documentos
type ProcessRelationships struct {
	QuotationToSO    []DocumentLink `json:"quotation_to_so,omitempty"`
	SOToPO           []DocumentLink `json:"so_to_po,omitempty"`
	POToDelivery     []DocumentLink `json:"po_to_delivery,omitempty"`
	SOToDelivery     []DocumentLink `json:"so_to_delivery,omitempty"`
	SOToInvoice      []DocumentLink `json:"so_to_invoice,omitempty"`
	InvoiceToPayment []DocumentLink `json:"invoice_to_payment,omitempty"`
}

// DocumentLink representa um link entre documentos
type DocumentLink struct {
	FromDocumentID   int       `json:"from_document_id"`
	FromDocumentNo   string    `json:"from_document_no"`
	FromDocumentType string    `json:"from_document_type"`
	ToDocumentID     int       `json:"to_document_id"`
	ToDocumentNo     string    `json:"to_document_no"`
	ToDocumentType   string    `json:"to_document_type"`
	LinkDate         time.Time `json:"link_date"`
}

// ProfitabilityAnalysis representa análise de lucratividade
type ProfitabilityAnalysis struct {
	TotalRevenue    float64                 `json:"total_revenue"`
	TotalCosts      float64                 `json:"total_costs"`
	TotalProfit     float64                 `json:"total_profit"`
	ProfitMargin    float64                 `json:"profit_margin_percentage"`
	ROI             float64                 `json:"roi_percentage"`
	ByProduct       []ProductProfitability  `json:"by_product"`
	ByCustomer      []CustomerProfitability `json:"by_customer"`
	ByPeriod        []PeriodProfitability   `json:"by_period"`
	TopProfitable   []ProcessSummary        `json:"top_profitable_processes"`
	LeastProfitable []ProcessSummary        `json:"least_profitable_processes"`
}

// ProductProfitability representa lucratividade por produto
type ProductProfitability struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Revenue     float64 `json:"revenue"`
	Cost        float64 `json:"cost"`
	Profit      float64 `json:"profit"`
	Margin      float64 `json:"margin_percentage"`
	Quantity    int     `json:"quantity_sold"`
	OrderCount  int     `json:"order_count"`
}

// CustomerProfitability representa lucratividade por cliente
type CustomerProfitability struct {
	ContactID     int     `json:"contact_id"`
	ContactName   string  `json:"contact_name"`
	ContactType   string  `json:"contact_type"`
	Revenue       float64 `json:"revenue"`
	Cost          float64 `json:"cost"`
	Profit        float64 `json:"profit"`
	Margin        float64 `json:"margin_percentage"`
	ProcessCount  int     `json:"process_count"`
	LifetimeValue float64 `json:"lifetime_value"`
	AverageValue  float64 `json:"average_order_value"`
}

// PeriodProfitability representa lucratividade por período
type PeriodProfitability struct {
	Period     string  `json:"period"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	Revenue    float64 `json:"revenue"`
	Cost       float64 `json:"cost"`
	Profit     float64 `json:"profit"`
	Margin     float64 `json:"margin_percentage"`
	OrderCount int     `json:"order_count"`
	Growth     float64 `json:"growth_percentage"`
}

// ProcessSummary representa um resumo de processo
type ProcessSummary struct {
	ProcessID    int       `json:"process_id"`
	ContactName  string    `json:"contact_name"`
	Status       string    `json:"status"`
	TotalValue   float64   `json:"total_value"`
	Profit       float64   `json:"profit"`
	ProfitMargin float64   `json:"profit_margin_percentage"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date,omitempty"`
	Duration     int       `json:"duration_days,omitempty"`
}

// ProcessAnalytics representa análises do processo
type ProcessAnalytics struct {
	ProcessID         int                `json:"process_id"`
	Performance       ProcessPerformance `json:"performance"`
	Bottlenecks       []Bottleneck       `json:"bottlenecks"`
	Recommendations   []Recommendation   `json:"recommendations"`
	ForecastedOutcome Forecast           `json:"forecasted_outcome"`
}

// ProcessPerformance representa performance do processo
type ProcessPerformance struct {
	EfficiencyScore       float64 `json:"efficiency_score"`
	CompletionTime        int     `json:"completion_time_days"`
	AverageCompletionTime int     `json:"average_completion_time_days"`
	DelayedStages         int     `json:"delayed_stages"`
	CostEfficiency        float64 `json:"cost_efficiency_percentage"`
}

// Bottleneck representa um gargalo no processo
type Bottleneck struct {
	Stage           string `json:"stage"`
	Description     string `json:"description"`
	Impact          string `json:"impact"`
	Severity        string `json:"severity"`
	SuggestedAction string `json:"suggested_action"`
}

// Recommendation representa uma recomendação
type Recommendation struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Impact      string `json:"expected_impact"`
	Effort      string `json:"effort_required"`
}

// Forecast representa uma previsão
type Forecast struct {
	CompletionDate     time.Time `json:"completion_date"`
	EstimatedRevenue   float64   `json:"estimated_revenue"`
	EstimatedProfit    float64   `json:"estimated_profit"`
	ProbabilitySuccess float64   `json:"probability_success_percentage"`
	RiskFactors        []string  `json:"risk_factors"`
}
