package quotation

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"go.uber.org/zap"
)

// GetStats obtém estatísticas de cotações
func (s *Service) GetStats(ctx context.Context, dateRange *dto.DateRange) (*dto.QuotationStats, error) {
	s.logger.Info("Obtendo estatísticas de cotações", zap.Any("date_range", dateRange))

	// Obter todas as cotações
	result, err := s.quotationRepo.GetAllQuotations(&pagination.PaginationParams{
		Page:     FirstPage,
		PageSize: DefaultBatchSize,
	})

	if err != nil {
		s.logger.Error("Erro ao buscar cotações para estatísticas", zap.Error(err))
		return nil, fmt.Errorf("falha ao buscar cotações para estatísticas: %w", err)
	}

	quotations := s.extractQuotationsFromResult(result)

	// Inicializar estatísticas
	stats := &dto.QuotationStats{
		TotalCount:         0,
		TotalValue:         0,
		AverageValue:       0,
		CountByStatus:      make(map[string]int),
		TotalValueByStatus: make(map[string]float64),
		CountByContact:     make(map[int]int),
		ExpiryDistribution: make(map[string]int),
	}

	// Calcular estatísticas
	for _, quotation := range quotations {
		// Aplicar filtro de data se fornecido
		if dateRange != nil {
			if quotation.CreatedAt.Before(dateRange.StartDate) || quotation.CreatedAt.After(dateRange.EndDate) {
				continue
			}
		}

		stats.TotalCount++
		stats.TotalValue += quotation.GrandTotal
		stats.CountByStatus[quotation.Status]++
		stats.TotalValueByStatus[quotation.Status] += quotation.GrandTotal
		stats.CountByContact[quotation.ContactID]++

		// Distribuição por mês de expiração
		expiryMonth := quotation.ExpiryDate.Format(MonthYearFormat)
		stats.ExpiryDistribution[expiryMonth]++
	}

	// Calcular média
	if stats.TotalCount > 0 {
		stats.AverageValue = stats.TotalValue / float64(stats.TotalCount)
	}

	s.logger.Info("Estatísticas de cotações calculadas com sucesso",
		zap.Int("total_count", stats.TotalCount),
		zap.Float64("total_value", stats.TotalValue),
		zap.Float64("average_value", stats.AverageValue))
	return stats, nil
}

// GetConversionStats obtém estatísticas de conversão de cotações
func (s *Service) GetConversionStats(ctx context.Context, dateRange *dto.DateRange) (*dto.ConversionRateStats, error) {
	s.logger.Info("Obtendo estatísticas de conversão de cotações", zap.Any("date_range", dateRange))

	// Obter todas as cotações
	result, err := s.quotationRepo.GetAllQuotations(&pagination.PaginationParams{
		Page:     FirstPage,
		PageSize: DefaultBatchSize,
	})

	if err != nil {
		s.logger.Error("Erro ao buscar cotações para estatísticas de conversão", zap.Error(err))
		return nil, fmt.Errorf("falha ao buscar cotações para estatísticas de conversão: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Inicializar estatísticas
	stats := &dto.ConversionRateStats{
		TotalQuotations:      0,
		ConvertedQuotations:  0,
		ConversionRate:       0,
		AverageTimeToConvert: 0,
		ConversionByContact:  make(map[int]float64),
		ValueConversionRate:  0,
	}

	var totalDays float64
	var convertedWithTime int
	var totalValue float64
	var convertedValue float64

	// Calcular estatísticas
	for _, quotation := range quotations {
		// Aplicar filtro de data se fornecido
		if dateRange != nil {
			if quotation.CreatedAt.Before(dateRange.StartDate) || quotation.CreatedAt.After(dateRange.EndDate) {
				continue
			}
		}

		stats.TotalQuotations++
		totalValue += quotation.GrandTotal

		// Verificar se a cotação foi convertida em pedido de venda
		if quotation.Status == models.QuotationStatusAccepted {
			salesOrder, err := s.salesOrderRepo.GetSalesOrdersByQuotation(quotation.ID)
			if err == nil && salesOrder != nil {
				stats.ConvertedQuotations++
				convertedValue += salesOrder.GrandTotal

				// Calcular tempo de conversão em dias
				daysToConvert := salesOrder.CreatedAt.Sub(quotation.CreatedAt).Hours() / HoursPerDay
				totalDays += daysToConvert
				convertedWithTime++
			}
		}
	}

	// Calcular taxas
	if stats.TotalQuotations > 0 {
		stats.ConversionRate = (float64(stats.ConvertedQuotations) / float64(stats.TotalQuotations)) * 100
	}

	if totalValue > 0 {
		stats.ValueConversionRate = (convertedValue / totalValue) * 100
	}

	// Calcular tempo médio de conversão
	if convertedWithTime > 0 {
		stats.AverageTimeToConvert = int(totalDays / float64(convertedWithTime))
	}

	s.logger.Info("Estatísticas de conversão de cotações calculadas com sucesso",
		zap.Int("total_quotations", stats.TotalQuotations),
		zap.Int("converted_quotations", stats.ConvertedQuotations),
		zap.Float64("conversion_rate", stats.ConversionRate))
	return stats, nil
}
