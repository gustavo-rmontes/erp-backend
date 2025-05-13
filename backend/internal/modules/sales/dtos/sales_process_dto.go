package dtos

import "time"

// SalesProcessCreateDTO representa os dados para criar um sales process
type SalesProcessCreateDTO struct {
	ContactID    int     `json:"contact_id" validate:"required"`
	Notes        string  `json:"notes,omitempty"`
	InitialValue float64 `json:"initial_value,omitempty"`
}

// SalesProcessUpdateDTO representa os dados para atualizar um sales process
type SalesProcessUpdateDTO struct {
	Notes      *string  `json:"notes,omitempty"`
	TotalValue *float64 `json:"total_value,omitempty"`
	Profit     *float64 `json:"profit,omitempty"`
}

// SalesProcessResponseDTO representa os dados retornados de um sales process
type SalesProcessResponseDTO struct {
	ID                 int                `json:"id"`
	ContactID          int                `json:"contact_id"`
	Contact            *ContactBasicInfo  `json:"contact,omitempty"`
	Status             string             `json:"status"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	TotalValue         float64            `json:"total_value"`
	TotalCost          float64            `json:"total_cost"`
	Profit             float64            `json:"profit"`
	ProfitMargin       float64            `json:"profit_margin"`
	Notes              string             `json:"notes,omitempty"`
	CurrentStage       string             `json:"current_stage"`
	CompletionRate     float64            `json:"completion_rate"`
	EstimatedCloseDate *time.Time         `json:"estimated_close_date,omitempty"`
	ActualCloseDate    *time.Time         `json:"actual_close_date,omitempty"`
	CycleTime          int                `json:"cycle_time_days,omitempty"`
	LinkedDocuments    LinkedDocumentsDTO `json:"linked_documents"`
}

// SalesProcessListItemDTO representa uma versão resumida para listagens
type SalesProcessListItemDTO struct {
	ID             int               `json:"id"`
	ContactID      int               `json:"contact_id"`
	Contact        *ContactBasicInfo `json:"contact,omitempty"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	TotalValue     float64           `json:"total_value"`
	Profit         float64           `json:"profit"`
	ProfitMargin   float64           `json:"profit_margin"`
	CurrentStage   string            `json:"current_stage"`
	CompletionRate float64           `json:"completion_rate"`
	LastActivity   time.Time         `json:"last_activity"`
}

// ProcessStatusUpdateDTO representa dados para atualizar status
type ProcessStatusUpdateDTO struct {
	Status string `json:"status" validate:"required"`
	Notes  string `json:"notes,omitempty"`
}

// LinkDocumentDTO representa dados para vincular documento
type LinkDocumentDTO struct {
	DocumentType string `json:"document_type" validate:"required,oneof=quotation sales_order purchase_order delivery invoice payment"`
	DocumentID   int    `json:"document_id" validate:"required"`
	Notes        string `json:"notes,omitempty"`
}

// UnlinkDocumentDTO representa dados para desvincular documento
type UnlinkDocumentDTO struct {
	DocumentType string `json:"document_type" validate:"required"`
	DocumentID   int    `json:"document_id" validate:"required"`
	Reason       string `json:"reason,omitempty"`
}

// ProcessStageTransitionDTO representa transição de estágio
type ProcessStageTransitionDTO struct {
	FromStage string    `json:"from_stage"`
	ToStage   string    `json:"to_stage" validate:"required"`
	Timestamp time.Time `json:"timestamp"`
	Notes     string    `json:"notes,omitempty"`
	UserID    int       `json:"user_id,omitempty"`
}

// LinkedDocumentsDTO representa documentos vinculados
type LinkedDocumentsDTO struct {
	Quotations     []LinkedDocumentInfo `json:"quotations,omitempty"`
	SalesOrders    []LinkedDocumentInfo `json:"sales_orders,omitempty"`
	PurchaseOrders []LinkedDocumentInfo `json:"purchase_orders,omitempty"`
	Deliveries     []LinkedDocumentInfo `json:"deliveries,omitempty"`
	Invoices       []LinkedDocumentInfo `json:"invoices,omitempty"`
	Payments       []LinkedDocumentInfo `json:"payments,omitempty"`
}

// LinkedDocumentInfo representa informação de documento vinculado
type LinkedDocumentInfo struct {
	ID         int       `json:"id"`
	DocumentNo string    `json:"document_no"`
	Status     string    `json:"status"`
	Value      float64   `json:"value,omitempty"`
	Date       time.Time `json:"date"`
	LinkedAt   time.Time `json:"linked_at"`
	LinkedBy   string    `json:"linked_by,omitempty"`
}

// ProcessMetricsDTO representa métricas do processo
type ProcessMetricsDTO struct {
	ProcessID           int                `json:"process_id"`
	CycleTime           int                `json:"cycle_time_days"`
	StagesDuration      []StageDurationDTO `json:"stages_duration"`
	ConversionRate      float64            `json:"conversion_rate"`
	AverageResponseTime float64            `json:"average_response_time_hours"`
	TouchPoints         int                `json:"touch_points"`
	DocumentsGenerated  int                `json:"documents_generated"`
}

// StageDurationDTO representa duração de estágio
type StageDurationDTO struct {
	Stage     string     `json:"stage"`
	Duration  int        `json:"duration_days"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	IsActive  bool       `json:"is_active"`
}

// ProcessAutomationDTO representa automação do processo
type ProcessAutomationDTO struct {
	ProcessID         int                 `json:"process_id" validate:"required"`
	AutoCreatePO      bool                `json:"auto_create_po"`
	AutoCreateInvoice bool                `json:"auto_create_invoice"`
	AutoSendEmails    bool                `json:"auto_send_emails"`
	FollowUpDays      int                 `json:"follow_up_days"`
	AlertOnDelay      bool                `json:"alert_on_delay"`
	EscalationRules   []EscalationRuleDTO `json:"escalation_rules,omitempty"`
}

// EscalationRuleDTO representa regra de escalonamento
type EscalationRuleDTO struct {
	TriggerDays   int    `json:"trigger_days"`
	Action        string `json:"action"`
	NotifyUserIDs []int  `json:"notify_user_ids"`
	EmailTemplate string `json:"email_template,omitempty"`
}

// ProcessTemplateDTO representa template de processo
type ProcessTemplateDTO struct {
	Name        string               `json:"name" validate:"required"`
	Description string               `json:"description,omitempty"`
	Stages      []StageTemplate      `json:"stages" validate:"required,min=1"`
	Documents   []string             `json:"documents"`
	Automations ProcessAutomationDTO `json:"automations"`
	IsActive    bool                 `json:"is_active"`
}

// StageTemplate representa template de estágio
type StageTemplate struct {
	Name         string   `json:"name" validate:"required"`
	Order        int      `json:"order" validate:"required,min=1"`
	Duration     int      `json:"estimated_duration_days"`
	Requirements []string `json:"requirements,omitempty"`
	AutoComplete bool     `json:"auto_complete"`
}

// ProcessCloneDTO representa dados para clonar processo
type ProcessCloneDTO struct {
	SourceProcessID  int    `json:"source_process_id" validate:"required"`
	ContactID        int    `json:"contact_id,omitempty"`
	IncludeDocuments bool   `json:"include_documents"`
	IncludeHistory   bool   `json:"include_history"`
	Notes            string `json:"notes,omitempty"`
}
