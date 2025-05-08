package quotation

// Constantes configuráveis do sistema
const (
	// Prefixos para documentos
	PrefixQuotation  = "QT"
	PrefixSalesOrder = "SO"

	// Formatos de data
	DateFormatForCodes = "20060102"
	DateFormatForCSV   = "2006-01-02"
	MonthYearFormat    = "2006-01"

	// Limites e valores padrão
	SequenceModulo      = 10000 // Módulo para gerar sequência de números
	DefaultBatchSize    = 1000  // Tamanho padrão de lote para operações em massa
	DefaultPageSize     = 100   // Tamanho padrão de página para operações regulares
	DefaultExpiryMonths = 1     // Meses padrão para expiração de cotações (usado em AddDate(0, 1, 0))

	// Formatação
	SequenceFormat = "%s-%s-%04d" // Formato para números de documentos: [Prefixo]-[Data]-[Sequência]

	// Separadores e formatação de texto
	RejectionNoteFormat = "%s\n\nRejection reason: %s"
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
	PDFPlaceholderFormat = "PDF da cotação #%s" // Formato para conteúdo de PDF placeholder
)

// Status de validação e regras de negócio
const (
	UpdateQuotationOnConversion = false // Define se a cotação deve ser atualizada após conversão
)
