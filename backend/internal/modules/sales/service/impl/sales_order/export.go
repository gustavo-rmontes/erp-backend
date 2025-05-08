package sales_order

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// GeneratePDF gera um PDF de um pedido de venda
func (s *Service) GeneratePDF(ctx context.Context, id int) ([]byte, error) {
	s.logger.Info("Gerando PDF do pedido de venda", zap.Int("order_id", id))

	// Buscar pedido
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para gerar PDF", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido para gerar PDF: %w", err)
	}

	// Em uma implementação real, usaríamos uma biblioteca de geração de PDF
	// Como este é um exemplo, retornamos um placeholder
	pdfContent := fmt.Sprintf(PDFPlaceholderFormat, order.SONo)

	s.logger.Info("PDF do pedido gerado com sucesso", zap.Int("order_id", id))
	return []byte(pdfContent), nil
}

// ExportToCSV exporta pedidos para CSV
func (s *Service) ExportToCSV(ctx context.Context, ids []int) ([]byte, error) {
	s.logger.Info("Exportando pedidos para CSV", zap.Ints("order_ids", ids))

	var orders []*models.SalesOrder

	// Buscar cada pedido
	for _, id := range ids {
		order, err := s.salesOrderRepo.GetSalesOrderByID(id)
		if err != nil {
			s.logger.Warn("Erro ao buscar pedido para exportação CSV", zap.Error(err), zap.Int("order_id", id))
			continue
		}
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		s.logger.Error("Nenhum pedido encontrado para exportação CSV", zap.Ints("order_ids", ids))
		return nil, fmt.Errorf("nenhum pedido encontrado para exportação")
	}

	// Criar buffer para CSV
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// Escrever cabeçalho
	header := []string{
		"ID", "Número do Pedido", "Cliente", "Cotação", "Status", "Data de Criação",
		"Data Prevista", "Subtotal", "Impostos", "Descontos", "Total",
		"Endereço de Entrega", "Condições de Pagamento",
	}
	if err := writer.Write(header); err != nil {
		s.logger.Error("Erro ao escrever cabeçalho CSV", zap.Error(err))
		return nil, fmt.Errorf("falha ao escrever cabeçalho CSV: %w", err)
	}

	// Escrever dados de cada pedido
	for _, order := range orders {
		contactName := ""
		if order.Contact != nil {
			contactName = order.Contact.Name
		}

		quotationNo := ""
		if order.Quotation != nil {
			quotationNo = order.Quotation.QuotationNo
		}

		row := []string{
			strconv.Itoa(order.ID),
			order.SONo,
			contactName,
			quotationNo,
			order.Status,
			order.CreatedAt.Format(DateFormatForCSV),
			order.ExpectedDate.Format(DateFormatForCSV),
			strconv.FormatFloat(order.SubTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(order.TaxTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(order.DiscountTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(order.GrandTotal, 'f', DecimalPrecision, 64),
			order.ShippingAddress,
			order.PaymentTerms,
		}

		if err := writer.Write(row); err != nil {
			s.logger.Error("Erro ao escrever linha CSV", zap.Error(err), zap.Int("order_id", order.ID))
			return nil, fmt.Errorf("falha ao escrever linha CSV: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		s.logger.Error("Erro ao finalizar CSV", zap.Error(err))
		return nil, fmt.Errorf("falha ao finalizar CSV: %w", err)
	}

	s.logger.Info("Pedidos exportados para CSV com sucesso", zap.Int("order_count", len(orders)))
	return buf.Bytes(), nil
}
