// stats.go (função otimizada)
package sales_order

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// GetStats obtém estatísticas de pedidos de venda
func (s *Service) GetStats(ctx context.Context, dateRange *dto.DateRange) (*dto.SalesOrderStats, error) {
	s.logger.Info("Obtendo estatísticas de pedidos de venda", zap.Any("date_range", dateRange))

	// Inicializar estatísticas
	stats := &dto.SalesOrderStats{
		TotalCount:         0,
		TotalValue:         0,
		AverageOrderValue:  0,
		CountByStatus:      make(map[string]int),
		TotalValueByStatus: make(map[string]float64),
		CountByContact:     make(map[int]int),
		TopContacts:        []dto.ContactStatsItem{},
		TopProducts:        []dto.ProductStatsItem{},
	}

	// Preparar a query base
	db := s.salesOrderRepo.GetDB().Model(&models.SalesOrder{})

	// Aplicar filtro de data se fornecido
	if dateRange != nil {
		db = db.Where("created_at BETWEEN ? AND ?", dateRange.StartDate, dateRange.EndDate)
	}

	// 1. Obter contagem total e soma total
	type TotalStat struct {
		Count int64
		Sum   float64
	}
	var totalStat TotalStat
	if err := db.Select("COUNT(*) as count, COALESCE(SUM(grand_total), 0) as sum").Scan(&totalStat).Error; err != nil {
		s.logger.Error("Erro ao calcular estatísticas totais", zap.Error(err))
		return nil, fmt.Errorf("falha ao calcular estatísticas totais: %w", err)
	}

	stats.TotalCount = int(totalStat.Count)
	stats.TotalValue = totalStat.Sum

	// Calcular média se houver pedidos
	if stats.TotalCount > 0 {
		stats.AverageOrderValue = stats.TotalValue / float64(stats.TotalCount)
	}

	// 2. Obter contagens e valores por status
	type StatusStat struct {
		Status string
		Count  int64
		Sum    float64
	}

	var statusStats []StatusStat
	if err := db.Select("status, COUNT(*) as count, COALESCE(SUM(grand_total), 0) as sum").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		s.logger.Error("Erro ao calcular estatísticas por status", zap.Error(err))
		return nil, fmt.Errorf("falha ao calcular estatísticas por status: %w", err)
	}

	for _, stat := range statusStats {
		stats.CountByStatus[stat.Status] = int(stat.Count)
		stats.TotalValueByStatus[stat.Status] = stat.Sum
	}

	// 3. Obter contagens por contato
	type ContactStat struct {
		ContactID int
		Count     int64
	}

	var contactStats []ContactStat
	if err := db.Select("contact_id, COUNT(*) as count").
		Group("contact_id").
		Scan(&contactStats).Error; err != nil {
		s.logger.Error("Erro ao calcular estatísticas por contato", zap.Error(err))
		return nil, fmt.Errorf("falha ao calcular estatísticas por contato: %w", err)
	}

	for _, stat := range contactStats {
		stats.CountByContact[stat.ContactID] = int(stat.Count)
	}

	// 4. Obter top contatos (com nome)
	topContactCount := 5 // Número de top contatos a retornar
	var topContacts []struct {
		ContactID int
		Name      string
		Count     int64
		Sum       float64
	}

	if err := db.Table("sales_orders").
		Select("sales_orders.contact_id, contacts.name, COUNT(*) as count, COALESCE(SUM(sales_orders.grand_total), 0) as sum").
		Joins("LEFT JOIN contacts ON contacts.id = sales_orders.contact_id").
		Group("sales_orders.contact_id, contacts.name").
		Order("count DESC").
		Limit(topContactCount).
		Scan(&topContacts).Error; err != nil {
		s.logger.Error("Erro ao obter top contatos", zap.Error(err))
		// Não interrompe o fluxo, apenas registra o erro
	} else {
		// Preencher os top contatos - com os nomes de campo corretos
		for _, contact := range topContacts {
			stats.TopContacts = append(stats.TopContacts, dto.ContactStatsItem{
				ContactID:   contact.ContactID,
				Name:        contact.Name,
				TotalValue:  contact.Sum,
				OrdersCount: int(contact.Count), // Nome correto: OrdersCount, não OrderCount
			})
		}
	}

	// 5. Obter top produtos
	topProductCount := 5 // Número de top produtos a retornar
	var topProducts []struct {
		ProductID   int
		ProductName string
		Count       int64
		Quantity    int64
	}

	if err := db.Table("so_items").
		Select("so_items.product_id, so_items.product_name, COUNT(DISTINCT so_items.sales_order_id) as count, SUM(so_items.quantity) as quantity").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = so_items.sales_order_id").
		Where("sales_orders.id IS NOT NULL").
		Group("so_items.product_id, so_items.product_name").
		Order("count DESC").
		Limit(topProductCount).
		Scan(&topProducts).Error; err != nil {
		s.logger.Error("Erro ao obter top produtos", zap.Error(err))
		// Não interrompe o fluxo, apenas registra o erro
	} else {
		// Preencher os top produtos - com os nomes de campo corretos
		for _, product := range topProducts {
			stats.TopProducts = append(stats.TopProducts, dto.ProductStatsItem{
				ProductID:   product.ProductID,
				ProductName: product.ProductName,
				OrdersCount: int(product.Count),
				Quantity:    int(product.Quantity),
			})
		}
	}

	s.logger.Info("Estatísticas de pedidos calculadas com sucesso",
		zap.Int("total_count", stats.TotalCount),
		zap.Float64("total_value", stats.TotalValue),
		zap.Float64("average_value", stats.AverageOrderValue))

	return stats, nil
}
