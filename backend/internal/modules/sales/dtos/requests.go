package dtos

// BulkUpdateRequest representa uma requisição de atualização em massa
type BulkUpdateRequest struct {
	IDs          []int                  `json:"ids" validate:"required,min=1"`
	UpdateData   map[string]interface{} `json:"update_data" validate:"required"`
	UpdateFields []string               `json:"update_fields,omitempty"`
	DryRun       bool                   `json:"dry_run"`
}

// BulkDeleteRequest representa uma requisição de exclusão em massa
type BulkDeleteRequest struct {
	IDs           []int  `json:"ids" validate:"required,min=1"`
	Force         bool   `json:"force"`
	ConfirmDelete bool   `json:"confirm_delete" validate:"required"`
	Reason        string `json:"reason,omitempty"`
}

// DateRangeRequest representa uma requisição de intervalo de datas
type DateRangeRequest struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
	Timezone  string `json:"timezone,omitempty"`
}

// SearchRequest representa uma requisição de busca genérica
type SearchRequest struct {
	Query         string            `json:"query" validate:"required"`
	Filters       map[string]string `json:"filters,omitempty"`
	Sort          string            `json:"sort,omitempty"`
	Order         string            `json:"order,omitempty" validate:"omitempty,oneof=asc desc"`
	Page          int               `json:"page" validate:"min=1"`
	PageSize      int               `json:"page_size" validate:"min=1,max=100"`
	IncludeNested bool              `json:"include_nested"`
}

// ExportRequest representa uma requisição de exportação
type ExportRequest struct {
	Format     string            `json:"format" validate:"required,oneof=csv excel pdf json xml"`
	Filters    map[string]string `json:"filters,omitempty"`
	Fields     []string          `json:"fields,omitempty"`
	DateRange  *DateRangeRequest `json:"date_range,omitempty"`
	IncludeAll bool              `json:"include_all"`
	Compress   bool              `json:"compress"`
	EmailTo    string            `json:"email_to,omitempty" validate:"omitempty,email"`
}

// ImportRequest representa uma requisição de importação
type ImportRequest struct {
	Format         string            `json:"format" validate:"required,oneof=csv excel json xml"`
	MappingRules   map[string]string `json:"mapping_rules,omitempty"`
	ValidateOnly   bool              `json:"validate_only"`
	UpdateExisting bool              `json:"update_existing"`
	SkipErrors     bool              `json:"skip_errors"`
}

// BulkStatusUpdateRequest representa atualização de status em massa
type BulkStatusUpdateRequest struct {
	IDs         []int  `json:"ids" validate:"required,min=1"`
	Status      string `json:"status" validate:"required"`
	Reason      string `json:"reason,omitempty"`
	NotifyUsers bool   `json:"notify_users"`
}

// MergeRequest representa uma requisição de merge
type MergeRequest struct {
	SourceIDs  []int             `json:"source_ids" validate:"required,min=2"`
	TargetID   int               `json:"target_id,omitempty"`
	MergeRules map[string]string `json:"merge_rules,omitempty"`
	Preview    bool              `json:"preview"`
}

// DuplicateCheckRequest representa verificação de duplicatas
type DuplicateCheckRequest struct {
	Fields      []string `json:"fields" validate:"required,min=1"`
	Threshold   float64  `json:"threshold,omitempty" validate:"omitempty,min=0,max=1"`
	IncludeNear bool     `json:"include_near_matches"`
	MaxResults  int      `json:"max_results,omitempty" validate:"omitempty,min=1,max=100"`
}

// BatchRequest representa uma requisição em lote
type BatchRequest struct {
	Operations  []BatchOperation `json:"operations" validate:"required,min=1,max=100"`
	Atomic      bool             `json:"atomic"`
	StopOnError bool             `json:"stop_on_error"`
}

// BatchOperation representa uma operação em lote
type BatchOperation struct {
	ID      string                 `json:"id" validate:"required"`
	Method  string                 `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	Path    string                 `json:"path" validate:"required"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
}

// FilterGroupRequest representa grupo de filtros complexos
type FilterGroupRequest struct {
	Operator string               `json:"operator" validate:"required,oneof=AND OR"`
	Filters  []FilterCondition    `json:"filters" validate:"required,min=1"`
	Groups   []FilterGroupRequest `json:"groups,omitempty"`
}

// FilterCondition representa uma condição de filtro
type FilterCondition struct {
	Field    string      `json:"field" validate:"required"`
	Operator string      `json:"operator" validate:"required,oneof=eq ne gt gte lt lte in nin contains starts_with ends_with"`
	Value    interface{} `json:"value" validate:"required"`
}

// AggregationRequest representa requisição de agregação
type AggregationRequest struct {
	GroupBy []string            `json:"group_by" validate:"required,min=1"`
	Metrics []AggregationMetric `json:"metrics" validate:"required,min=1"`
	Filters []FilterCondition   `json:"filters,omitempty"`
	Having  []FilterCondition   `json:"having,omitempty"`
	Sort    string              `json:"sort,omitempty"`
	Order   string              `json:"order,omitempty" validate:"omitempty,oneof=asc desc"`
	Limit   int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// AggregationMetric representa uma métrica de agregação
type AggregationMetric struct {
	Field    string `json:"field" validate:"required"`
	Function string `json:"function" validate:"required,oneof=count sum avg min max"`
	Alias    string `json:"alias,omitempty"`
}

// WebhookRequest representa configuração de webhook
type WebhookRequest struct {
	URL         string            `json:"url" validate:"required,url"`
	Events      []string          `json:"events" validate:"required,min=1"`
	Headers     map[string]string `json:"headers,omitempty"`
	Secret      string            `json:"secret,omitempty"`
	Active      bool              `json:"active"`
	RetryPolicy RetryPolicyDTO    `json:"retry_policy,omitempty"`
}

// RetryPolicyDTO representa política de retry
type RetryPolicyDTO struct {
	MaxAttempts   int     `json:"max_attempts" validate:"min=1,max=10"`
	InitialDelay  int     `json:"initial_delay_seconds" validate:"min=1"`
	MaxDelay      int     `json:"max_delay_seconds" validate:"min=1"`
	BackoffFactor float64 `json:"backoff_factor" validate:"min=1,max=5"`
}
