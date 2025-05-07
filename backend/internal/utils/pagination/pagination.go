package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// PaginationParams contém os parâmetros para paginação
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginatedResult contém o resultado paginado
type PaginatedResult struct {
	TotalItems  int64
	TotalPages  int
	CurrentPage int
	PageSize    int
	Items       any
}

// DefaultPage é o número de página padrão
const DefaultPage = 1

// DefaultPageSize é o tamanho de página padrão
const DefaultPageSize = 10

// MaxPageSize é o tamanho máximo de página permitido
const MaxPageSize = 100

// NewPaginationParams cria um novo parâmetro de paginação a partir de uma requisição HTTP
func NewPaginationParams(r *http.Request) PaginationParams {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}

	// Limita o tamanho da página
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// NewPaginatedResult cria um novo resultado paginado
func NewPaginatedResult(totalItems int64, page, pageSize int, items interface{}) *PaginatedResult {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return &PaginatedResult{
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		Items:       items,
	}
}

// CalculateOffset calcula o offset para a consulta SQL
func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// Validate valida os parâmetros de paginação
func (p *PaginationParams) Validate() bool {
	return p.Page > 0 && p.PageSize > 0
}
