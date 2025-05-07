// Package dto contém todos os Data Transfer Objects (DTOs) do sistema ERP.
// Os DTOs servem como contratos entre as diferentes camadas da aplicação,
// especialmente entre a API REST e os serviços de negócio.
// Versão 1.0.0
package dto

import "time"

// Version representa a versão atual dos DTOs.
// Útil para controle de compatibilidade em APIs e documentação.
const Version = "1.0.0"

// Pagination contém informações comuns de paginação usadas em todas as respostas paginadas.
// Esta estrutura substitui a definição duplicada em cada tipo de resposta paginada.
type Pagination struct {
	TotalItems  int64 `json:"total_items"`  // Total de itens encontrados
	TotalPages  int   `json:"total_pages"`  // Total de páginas disponíveis
	CurrentPage int   `json:"current_page"` // Página atual
	PageSize    int   `json:"page_size"`    // Tamanho da página
}

// PaginatedResponse é uma estrutura genérica para respostas paginadas.
// Utiliza generics para suportar qualquer tipo de item nos resultados.
type PaginatedResponse[T any] struct {
	Items []T `json:"items"` // Lista de itens do tipo específico
	Pagination
}

// BaseItem contém os campos comuns a todos os itens de documentos (pedidos, cotações, etc).
// Esta estrutura reduz a duplicação entre diferentes tipos de itens.
type BaseItem struct {
	ProductID   int     `json:"product_id" validate:"required"`      // ID do produto
	ProductName string  `json:"product_name" validate:"required"`    // Nome do produto
	Quantity    int     `json:"quantity" validate:"required,gt=0"`   // Quantidade
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"` // Preço unitário
	Discount    float64 `json:"discount" validate:"min=0"`           // Desconto em porcentagem
	Tax         float64 `json:"tax" validate:"min=0"`                // Taxa em porcentagem
	Description string  `json:"description,omitempty"`               // Descrição opcional
}

// BaseItemResponse contém os campos comuns a todas as respostas de itens.
type BaseItemResponse struct {
	ID          int     `json:"id"`                     // ID do item
	ProductID   int     `json:"product_id"`             // ID do produto
	ProductName string  `json:"product_name"`           // Nome do produto
	ProductCode string  `json:"product_code,omitempty"` // Código do produto
	Description string  `json:"description,omitempty"`  // Descrição
	Quantity    int     `json:"quantity"`               // Quantidade
	UnitPrice   float64 `json:"unit_price"`             // Preço unitário
	Discount    float64 `json:"discount"`               // Desconto em porcentagem
	Tax         float64 `json:"tax"`                    // Taxa em porcentagem
	Total       float64 `json:"total"`                  // Valor total do item (somente leitura)
}

// BaseDocumentFields contém campos comuns a todos os documentos (pedidos, cotações, etc).
// Apenas para uso interno nas estruturas, não exposto diretamente.
type BaseDocumentFields struct {
	CreatedAt     time.Time `json:"created_at"`     // Data de criação (somente leitura)
	UpdatedAt     time.Time `json:"updated_at"`     // Data de atualização (somente leitura)
	SubTotal      float64   `json:"subtotal"`       // Subtotal (somente leitura)
	TaxTotal      float64   `json:"tax_total"`      // Total de impostos (somente leitura)
	DiscountTotal float64   `json:"discount_total"` // Total de descontos (somente leitura)
	GrandTotal    float64   `json:"grand_total"`    // Total geral (somente leitura)
}

// StatusUpdateRequest é uma estrutura base para solicitações de atualização de status.
type StatusUpdateRequest struct {
	Status string `json:"status" validate:"required"` // Novo status
	Reason string `json:"reason,omitempty"`           // Razão opcional para a mudança
}

// DocumentFilter define filtros comuns para consulta de documentos.
type DocumentFilter struct {
	StartDate   time.Time `json:"start_date,omitempty" form:"startDate"`     // Data inicial
	EndDate     time.Time `json:"end_date,omitempty" form:"endDate"`         // Data final
	Status      []string  `json:"status,omitempty" form:"status"`            // Status do documento
	ContactID   int       `json:"contact_id,omitempty" form:"contactId"`     // ID do contato
	MinValue    float64   `json:"min_value,omitempty" form:"minValue"`       // Valor mínimo
	MaxValue    float64   `json:"max_value,omitempty" form:"maxValue"`       // Valor máximo
	SearchQuery string    `json:"search_query,omitempty" form:"searchQuery"` // Texto para busca
}

// DateRange representa um intervalo de datas para filtros e relatórios.
type DateRange struct {
	StartDate time.Time `json:"start_date" validate:"required"`                 // Data inicial
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"` // Data final
}

// EmailOptions contém opções para envio de e-mail.
type EmailOptions struct {
	To          []string `json:"to" validate:"required,min=1,dive,email"`       // Destinatários principais
	Cc          []string `json:"cc,omitempty" validate:"omitempty,dive,email"`  // Cópia
	Bcc         []string `json:"bcc,omitempty" validate:"omitempty,dive,email"` // Cópia oculta
	Subject     string   `json:"subject,omitempty"`                             // Assunto
	Message     string   `json:"message,omitempty"`                             // Mensagem
	AttachPDF   bool     `json:"attach_pdf"`                                    // Anexar PDF
	CustomTheme string   `json:"custom_theme,omitempty"`                        // Tema customizado
}

// APIError define uma estrutura padronizada para erros retornados pela API.
type APIError struct {
	Code    string `json:"code"`              // Código do erro
	Message string `json:"message"`           // Mensagem amigável
	Details string `json:"details,omitempty"` // Detalhes adicionais
}
