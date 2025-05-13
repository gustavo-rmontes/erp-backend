package dtos

import "time"

// SuccessResponse representa uma resposta de sucesso genérica
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	ErrorCode string    `json:"error_code,omitempty"`
	Details   []string  `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// ValidationErrorResponse representa erros de validação
type ValidationErrorResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Errors  []ValidationError    `json:"errors"`
	Meta    *ValidationErrorMeta `json:"meta,omitempty"`
}

// ValidationError representa um erro de validação individual
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// ValidationErrorMeta representa metadados do erro de validação
type ValidationErrorMeta struct {
	TotalErrors int      `json:"total_errors"`
	Fields      []string `json:"fields"`
}

// BulkOperationResponse representa resposta de operação em massa
type BulkOperationResponse struct {
	Success        bool             `json:"success"`
	TotalRequested int              `json:"total_requested"`
	TotalProcessed int              `json:"total_processed"`
	TotalSucceeded int              `json:"total_succeeded"`
	TotalFailed    int              `json:"total_failed"`
	Results        []BulkResultItem `json:"results"`
	Errors         []BulkErrorItem  `json:"errors,omitempty"`
}

// BulkResultItem representa resultado individual de operação em massa
type BulkResultItem struct {
	ID      interface{} `json:"id"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// BulkErrorItem representa erro individual em operação em massa
type BulkErrorItem struct {
	ID      interface{} `json:"id"`
	Error   string      `json:"error"`
	Details []string    `json:"details,omitempty"`
}

// ExportResponse representa resposta de exportação
type ExportResponse struct {
	Success     bool       `json:"success"`
	FileID      string     `json:"file_id"`
	FileName    string     `json:"file_name"`
	Format      string     `json:"format"`
	Size        int64      `json:"size_bytes"`
	DownloadURL string     `json:"download_url,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at,omitempty"`
	RecordCount int        `json:"record_count"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ImportResponse representa resposta de importação
type ImportResponse struct {
	Success       bool            `json:"success"`
	ImportID      string          `json:"import_id"`
	TotalRows     int             `json:"total_rows"`
	ProcessedRows int             `json:"processed_rows"`
	SuccessRows   int             `json:"success_rows"`
	ErrorRows     int             `json:"error_rows"`
	SkippedRows   int             `json:"skipped_rows"`
	Status        string          `json:"status"`
	StartedAt     time.Time       `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Errors        []ImportError   `json:"errors,omitempty"`
	Warnings      []ImportWarning `json:"warnings,omitempty"`
}

// ImportError representa erro de importação
type ImportError struct {
	Row     int      `json:"row"`
	Column  string   `json:"column,omitempty"`
	Error   string   `json:"error"`
	Value   string   `json:"value,omitempty"`
	Details []string `json:"details,omitempty"`
}

// ImportWarning representa aviso de importação
type ImportWarning struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Warning string `json:"warning"`
	Value   string `json:"value,omitempty"`
}

// MetricsResponse representa resposta de métricas
type MetricsResponse struct {
	Success   bool                   `json:"success"`
	Period    string                 `json:"period"`
	StartDate time.Time              `json:"start_date"`
	EndDate   time.Time              `json:"end_date"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Charts    []ChartData            `json:"charts,omitempty"`
}

// MetricValue representa valor de métrica
type MetricValue struct {
	Value      interface{}     `json:"value"`
	Change     *float64        `json:"change_percentage,omitempty"`
	Trend      string          `json:"trend,omitempty"` // up, down, stable
	Comparison *ComparisonData `json:"comparison,omitempty"`
}

// ComparisonData representa dados de comparação
type ComparisonData struct {
	PreviousValue  interface{} `json:"previous_value"`
	PreviousPeriod string      `json:"previous_period"`
	Difference     float64     `json:"difference"`
	Percentage     float64     `json:"percentage"`
}

// ChartData representa dados para gráfico
type ChartData struct {
	Type    string                 `json:"type"` // line, bar, pie, etc
	Title   string                 `json:"title"`
	Series  []ChartSeries          `json:"series"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// ChartSeries representa série de dados do gráfico
type ChartSeries struct {
	Name  string        `json:"name"`
	Data  []interface{} `json:"data"`
	Color string        `json:"color,omitempty"`
}

// DashboardResponse representa resposta de dashboard
type DashboardResponse struct {
	Success    bool           `json:"success"`
	LastUpdate time.Time      `json:"last_update"`
	Widgets    []WidgetData   `json:"widgets"`
	Alerts     []AlertData    `json:"alerts,omitempty"`
	Activities []ActivityData `json:"activities,omitempty"`
}

// WidgetData representa dados de widget
type WidgetData struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Data     interface{}    `json:"data"`
	Config   interface{}    `json:"config,omitempty"`
	Position WidgetPosition `json:"position"`
}

// WidgetPosition representa posição do widget
type WidgetPosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AlertData representa dados de alerta
type AlertData struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	Severity string    `json:"severity"` // info, warning, error, critical
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Time     time.Time `json:"time"`
	Actions  []string  `json:"actions,omitempty"`
}

// ActivityData representa dados de atividade
type ActivityData struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	User        string    `json:"user"`
	Time        time.Time `json:"time"`
	Icon        string    `json:"icon,omitempty"`
	Link        string    `json:"link,omitempty"`
}

// MetaData representa metadados da resposta
type MetaData struct {
	RequestID    string          `json:"request_id"`
	ResponseTime int64           `json:"response_time_ms"`
	Version      string          `json:"version,omitempty"`
	Pagination   *PaginationMeta `json:"pagination,omitempty"`
}

// PaginationMeta representa metadados de paginação
type PaginationMeta struct {
	Page         int  `json:"page"`
	PageSize     int  `json:"page_size"`
	TotalPages   int  `json:"total_pages"`
	TotalRecords int  `json:"total_records"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
	NextPage     *int `json:"next_page,omitempty"`
	PreviousPage *int `json:"previous_page,omitempty"`
}

// NotificationResponse representa resposta de notificação
type NotificationResponse struct {
	Success        bool       `json:"success"`
	NotificationID string     `json:"notification_id"`
	Channel        string     `json:"channel"` // email, sms, push, webhook
	Recipients     []string   `json:"recipients"`
	DeliveredCount int        `json:"delivered_count"`
	FailedCount    int        `json:"failed_count"`
	Status         string     `json:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	ErrorDetails   []string   `json:"error_details,omitempty"`
}

// HealthCheckResponse representa resposta de health check
type HealthCheckResponse struct {
	Status    string                   `json:"status"` // healthy, degraded, unhealthy
	Version   string                   `json:"version"`
	Timestamp time.Time                `json:"timestamp"`
	Uptime    int64                    `json:"uptime_seconds"`
	Services  map[string]ServiceStatus `json:"services"`
	Metrics   HealthMetrics            `json:"metrics,omitempty"`
}

// ServiceStatus representa status de serviço
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency int64  `json:"latency_ms,omitempty"`
}

// HealthMetrics representa métricas de saúde
type HealthMetrics struct {
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage float64 `json:"memory_usage_percent"`
	DiskUsage   float64 `json:"disk_usage_percent"`
	Connections int     `json:"active_connections"`
}
