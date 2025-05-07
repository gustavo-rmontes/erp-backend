package errors

import (
	"errors"
)

// Erros comuns a todos os repositórios
var (
	// Erros de banco de dados
	ErrDatabaseConnection = errors.New("falha na conexão com o banco de dados")
	ErrTransactionFailed  = errors.New("falha na transação do banco de dados")

	// Erros de validação
	ErrInvalidPagination = errors.New("parâmetros de paginação inválidos")

	// Erros de entidade não encontrada
	ErrQuotationNotFound     = errors.New("cotação não encontrada")
	ErrSalesOrderNotFound    = errors.New("pedido de venda não encontrado")
	ErrPurchaseOrderNotFound = errors.New("pedido de compra não encontrado")
	ErrDeliveryNotFound      = errors.New("entrega não encontrada")
	ErrInvoiceNotFound       = errors.New("fatura não encontrada")
	ErrPaymentNotFound       = errors.New("pagamento não encontrado")
	ErrSalesProcessNotFound  = errors.New("processo de vendas não encontrado")

	// Erros de lógica de negócio
	ErrRelatedRecordsExist = errors.New("não é possível excluir devido a registros relacionados")
)

// WrapError adiciona um contexto a um erro
func WrapError(err error, message string) error {
	return errors.New(message + ": " + err.Error())
}

// IsNotFound verifica se o erro é do tipo "não encontrado"
func IsNotFound(err error) bool {
	return err == ErrQuotationNotFound ||
		err == ErrSalesOrderNotFound ||
		err == ErrPurchaseOrderNotFound ||
		err == ErrDeliveryNotFound ||
		err == ErrInvoiceNotFound ||
		err == ErrPaymentNotFound ||
		err == ErrSalesProcessNotFound
}
