// Package dto - módulo de DTOs para estatísticas e análises
// Este arquivo contém todas as estruturas de DTOs relacionadas a estatísticas,
// relatórios e análises dos diferentes módulos do sistema.
package dto

import "time"

// ContactStatsItem representa um item de estatística para contatos.
// Utilizado em diversas estatísticas que agrupam informações por contato.
type ContactStatsItem struct {
	ContactID   int     `json:"contact_id"`             // ID do contato
	Name        string  `json:"name"`                   // Nome do contato
	CompanyName string  `json:"company_name,omitempty"` // Nome da empresa (para PJ)
	Document    string  `json:"document"`               // Documento (CPF/CNPJ)
	TotalValue  float64 `json:"total_value"`            // Valor total movimentado
	OrdersCount int     `json:"orders_count"`           // Quantidade de pedidos
}

// ProductStatsItem representa um item de estatística para produtos.
// Utilizado em estatísticas que agrupam informações por produto.
type ProductStatsItem struct {
	ProductID   int     `json:"product_id"`             // ID do produto
	ProductName string  `json:"product_name"`           // Nome do produto
	ProductCode string  `json:"product_code,omitempty"` // Código do produto
	TotalValue  float64 `json:"total_value"`            // Valor total vendido
	Quantity    int     `json:"quantity"`               // Quantidade total vendida
	OrdersCount int     `json:"orders_count"`           // Número de pedidos que incluem o produto
}

// PeriodComparison compara dados entre períodos.
// Utilizado para comparar métricas entre períodos diferentes.
type PeriodComparison struct {
	CurrentPeriodValue  float64 `json:"current_period_value"`  // Valor do período atual
	PreviousPeriodValue float64 `json:"previous_period_value"` // Valor do período anterior
	PercentageChange    float64 `json:"percentage_change"`     // Variação percentual
}

// SalesOrderStats representa estatísticas resumidas de pedidos de venda.
type SalesOrderStats struct {
	TotalCount         int                `json:"total_count"`                 // Total de pedidos
	TotalValue         float64            `json:"total_value"`                 // Valor total dos pedidos
	AverageOrderValue  float64            `json:"average_order_value"`         // Valor médio por pedido
	CountByStatus      map[string]int     `json:"count_by_status"`             // Contagem por status
	TotalValueByStatus map[string]float64 `json:"total_value_by_status"`       // Valor total por status
	CountByContact     map[int]int        `json:"count_by_contact"`            // Contagem por contato
	TopContacts        []ContactStatsItem `json:"top_contacts"`                // Principais contatos
	TopProducts        []ProductStatsItem `json:"top_products"`                // Principais produtos
	PeriodComparison   *PeriodComparison  `json:"period_comparison,omitempty"` // Comparação com período anterior
}

// QuotationStats representa estatísticas resumidas de cotações.
type QuotationStats struct {
	TotalCount         int                `json:"total_count"`                 // Total de cotações
	TotalValue         float64            `json:"total_value"`                 // Valor total das cotações
	AverageValue       float64            `json:"average_value"`               // Valor médio por cotação
	CountByStatus      map[string]int     `json:"count_by_status"`             // Contagem por status
	TotalValueByStatus map[string]float64 `json:"total_value_by_status"`       // Valor total por status
	CountByContact     map[int]int        `json:"count_by_contact"`            // Contagem por contato
	ExpiryDistribution map[string]int     `json:"expiry_distribution"`         // Distribuição por mês de expiração
	TopContacts        []ContactStatsItem `json:"top_contacts"`                // Principais contatos
	PeriodComparison   *PeriodComparison  `json:"period_comparison,omitempty"` // Comparação com período anterior
}

// ConversionRateComparison compara taxas de conversão entre períodos.
type ConversionRateComparison struct {
	CurrentPeriodRate  float64 `json:"current_period_rate"`  // Taxa do período atual
	PreviousPeriodRate float64 `json:"previous_period_rate"` // Taxa do período anterior
	PercentageChange   float64 `json:"percentage_change"`    // Variação percentual
}

// ConversionRateStats representa estatísticas de conversão de cotações para pedidos.
type ConversionRateStats struct {
	TotalQuotations      int                       `json:"total_quotations"`            // Total de cotações
	ConvertedQuotations  int                       `json:"converted_quotations"`        // Cotações convertidas
	ConversionRate       float64                   `json:"conversion_rate"`             // Taxa de conversão (%)
	AverageTimeToConvert int                       `json:"average_time_to_convert"`     // Tempo médio de conversão (dias)
	ConversionByContact  map[int]float64           `json:"conversion_by_contact"`       // Taxa por contato
	ValueConversionRate  float64                   `json:"value_conversion_rate"`       // Taxa de valor convertido
	PeriodComparison     *ConversionRateComparison `json:"period_comparison,omitempty"` // Comparação com período anterior
}

// SalesProcessStats DTO para estatísticas do processo de vendas.
type SalesProcessStats struct {
	TotalProcesses     int                `json:"total_processes"`     // Total de processos
	ActiveProcesses    int                `json:"active_processes"`    // Processos ativos
	CompletedProcesses int                `json:"completed_processes"` // Processos concluídos
	TotalValue         float64            `json:"total_value"`         // Valor total
	TotalProfit        float64            `json:"total_profit"`        // Lucro total
	AverageProfit      float64            `json:"average_profit"`      // Lucro médio
	StatusCounts       map[string]int     `json:"status_counts"`       // Contagem por status
	TopContacts        []ContactStatsItem `json:"top_contacts"`        // Principais contatos
	MonthlyTrends      []MonthlyTrendItem `json:"monthly_trends"`      // Tendências mensais
}

// MonthlyTrendItem DTO para dados de tendência mensal.
type MonthlyTrendItem struct {
	Month      string  `json:"month"`       // Mês no formato "YYYY-MM"
	Count      int     `json:"count"`       // Contagem de itens
	TotalValue float64 `json:"total_value"` // Valor total
	Profit     float64 `json:"profit"`      // Lucro
}

// PaymentMethodSummary DTO para resumo de métodos de pagamento.
type PaymentMethodSummary struct {
	PaymentMethod string  `json:"payment_method"` // Método de pagamento
	Count         int     `json:"count"`          // Contagem de pagamentos
	TotalAmount   float64 `json:"total_amount"`   // Valor total
	Percentage    float64 `json:"percentage"`     // Percentual do total
}

// PaymentStatsResponse DTO para estatísticas de pagamentos.
type PaymentStatsResponse struct {
	TotalPaymentsCount   int                    `json:"total_payments_count"`   // Total de pagamentos
	TotalPaymentsAmount  float64                `json:"total_payments_amount"`  // Valor total de pagamentos
	MethodStats          []PaymentMethodSummary `json:"method_stats"`           // Estatísticas por método
	MonthlyPayments      []MonthlyPaymentStat   `json:"monthly_payments"`       // Pagamentos mensais
	AveragePaymentAmount float64                `json:"average_payment_amount"` // Valor médio de pagamento
	LargestPayment       float64                `json:"largest_payment"`        // Maior pagamento
	SmallestPayment      float64                `json:"smallest_payment"`       // Menor pagamento
}

// MonthlyPaymentStat DTO para estatísticas mensais de pagamento.
type MonthlyPaymentStat struct {
	Month  string  `json:"month"`  // Mês no formato "YYYY-MM"
	Count  int     `json:"count"`  // Contagem de pagamentos
	Amount float64 `json:"amount"` // Valor total
}

// InvoiceAgingReport DTO para relatórios de idade de faturas.
type InvoiceAgingReport struct {
	CurrentTotal     float64            `json:"current_total"`     // Total ainda não vencido
	OneDayTotal      float64            `json:"one_day_total"`     // Total 1-30 dias
	ThirtyDayTotal   float64            `json:"thirty_day_total"`  // Total 31-60 dias
	SixtyDayTotal    float64            `json:"sixty_day_total"`   // Total 61-90 dias
	NinetyDayTotal   float64            `json:"ninety_day_total"`  // Total 91+ dias
	TotalOutstanding float64            `json:"total_outstanding"` // Total geral em aberto
	Invoices         []InvoiceAgingItem `json:"invoices"`          // Lista de faturas
}

// InvoiceAgingItem DTO para item individual em relatório de idade de faturas.
type InvoiceAgingItem struct {
	ID          int       `json:"id"`           // ID da fatura
	InvoiceNo   string    `json:"invoice_no"`   // Número da fatura
	ContactID   int       `json:"contact_id"`   // ID do contato
	ContactName string    `json:"contact_name"` // Nome do contato
	DueDate     time.Time `json:"due_date"`     // Data de vencimento
	GrandTotal  float64   `json:"grand_total"`  // Valor total
	Balance     float64   `json:"balance"`      // Saldo em aberto
	DaysOverdue int       `json:"days_overdue"` // Dias de atraso
	AgingBucket string    `json:"aging_bucket"` // Faixa de atraso ("Current", "1-30", "31-60", "61-90", "91+")
}
