package sales_order

// Constantes configuráveis do sistema
const (
	// Prefixos para documentos
	PrefixSalesOrder    = "SO"
	PrefixPurchaseOrder = "PO"
	PrefixInvoice       = "INV"
	PrefixDelivery      = "DEL"

	// Formatos de data
	DateFormatForCodes = "20060102"
	DateFormatForCSV   = "2006-01-02"
	MonthYearFormat    = "2006-01"

	// Limites e valores padrão
	SequenceModulo   = 10000 // Módulo para gerar sequência de números
	DefaultBatchSize = 1000  // Tamanho padrão de lote para operações em massa
	DefaultPageSize  = 100   // Tamanho padrão de página para operações regulares

	// Formatação
	SequenceFormat = "%s-%s-%04d" // Formato para números de documentos: [Prefixo]-[Data]-[Sequência]
)

// Constantes numéricas
const (
	HoursPerDay       = 24    // Usado para converter horas em dias em cálculos de tempo
	PercentageDivisor = 100.0 // Usado em cálculos de percentual (desconto, imposto)
	DecimalPrecision  = 2     // Precisão padrão para valores monetários em CSV
	FirstPage         = 1     // Número da primeira página em paginações
)

// Mensagens padrão
const (
	PDFPlaceholderFormat = "PDF do pedido #%s" // Formato para conteúdo de PDF placeholder
)
