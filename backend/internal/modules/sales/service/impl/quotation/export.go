package quotation

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// GeneratePDF gera um PDF de uma cotação
func (s *Service) GeneratePDF(ctx context.Context, id int) ([]byte, error) {
	s.logger.Info("Gerando PDF da cotação", zap.Int("quotation_id", id))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para gerar PDF", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para gerar PDF: %w", err)
	}

	// Em uma implementação real, usaríamos uma biblioteca de geração de PDF
	// Como este é um exemplo, retornamos um placeholder
	pdfContent := fmt.Sprintf(PDFPlaceholderFormat, quotation.QuotationNo)

	s.logger.Info("PDF da cotação gerado com sucesso", zap.Int("quotation_id", id))
	return []byte(pdfContent), nil
}

// ExportToCSV exporta cotações para CSV
func (s *Service) ExportToCSV(ctx context.Context, ids []int) ([]byte, error) {
	s.logger.Info("Exportando cotações para CSV", zap.Ints("quotation_ids", ids))

	var quotations []*models.Quotation

	// Buscar cada cotação
	for _, id := range ids {
		quotation, err := s.quotationRepo.GetQuotationByID(id)
		if err != nil {
			s.logger.Warn("Erro ao buscar cotação para exportação CSV", zap.Error(err), zap.Int("quotation_id", id))
			continue
		}
		quotations = append(quotations, quotation)
	}

	if len(quotations) == 0 {
		s.logger.Error("Nenhuma cotação encontrada para exportação CSV", zap.Ints("quotation_ids", ids))
		return nil, fmt.Errorf("nenhuma cotação encontrada para exportação")
	}

	// Criar buffer para CSV
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// Escrever cabeçalho
	header := []string{
		"ID", "Número da Cotação", "Cliente", "Status", "Data de Criação",
		"Data de Expiração", "Subtotal", "Impostos", "Descontos", "Total",
	}
	if err := writer.Write(header); err != nil {
		s.logger.Error("Erro ao escrever cabeçalho CSV", zap.Error(err))
		return nil, fmt.Errorf("falha ao escrever cabeçalho CSV: %w", err)
	}

	// Escrever dados de cada cotação
	for _, quotation := range quotations {
		contactName := ""
		if quotation.Contact != nil {
			contactName = quotation.Contact.Name
		}

		row := []string{
			strconv.Itoa(quotation.ID),
			quotation.QuotationNo,
			contactName,
			quotation.Status,
			quotation.CreatedAt.Format(DateFormatForCSV),
			quotation.ExpiryDate.Format(DateFormatForCSV),
			strconv.FormatFloat(quotation.SubTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(quotation.TaxTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(quotation.DiscountTotal, 'f', DecimalPrecision, 64),
			strconv.FormatFloat(quotation.GrandTotal, 'f', DecimalPrecision, 64),
		}

		if err := writer.Write(row); err != nil {
			s.logger.Error("Erro ao escrever linha CSV", zap.Error(err), zap.Int("quotation_id", quotation.ID))
			return nil, fmt.Errorf("falha ao escrever linha CSV: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		s.logger.Error("Erro ao finalizar CSV", zap.Error(err))
		return nil, fmt.Errorf("falha ao finalizar CSV: %w", err)
	}

	s.logger.Info("Cotações exportadas para CSV com sucesso", zap.Int("quotation_count", len(quotations)))
	return buf.Bytes(), nil
}

// SendByEmail envia uma cotação por e-mail
func (s *Service) SendByEmail(ctx context.Context, id int, options *dto.EmailOptions) error {
	s.logger.Info("Enviando cotação por e-mail", zap.Int("quotation_id", id), zap.Any("options", options))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para envio de e-mail", zap.Error(err), zap.Int("quotation_id", id))
		return fmt.Errorf("falha ao buscar cotação para envio de e-mail: %w", err)
	}

	// Em uma implementação real, integraríamos com um serviço de e-mail
	// Como este é um exemplo, apenas logamos a ação
	s.logger.Info("Simulando envio de e-mail para cotação",
		zap.Int("quotation_id", id),
		zap.String("quotation_no", quotation.QuotationNo),
		zap.Strings("recipients", options.To),
		zap.String("subject", options.Subject))

	// Se a cotação está em draft, atualizar para enviada
	if quotation.Status == models.QuotationStatusDraft && options.AttachPDF {
		quotation.Status = models.QuotationStatusSent
		if err := s.quotationRepo.UpdateQuotation(id, quotation); err != nil {
			s.logger.Error("Erro ao atualizar status da cotação após envio", zap.Error(err), zap.Int("quotation_id", id))
			return fmt.Errorf("falha ao atualizar status da cotação após envio: %w", err)
		}
		s.logger.Info("Status da cotação atualizado para 'enviada'", zap.Int("quotation_id", id))
	}

	return nil
}
