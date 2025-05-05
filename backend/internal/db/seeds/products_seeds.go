package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"ERP-ONSMART/backend/internal/modules/products/models"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/lib/pq"
)

// SeedProducts gera produtos fictícios
func SeedProducts(db *sql.DB, count int) error {
	log.Printf("[seeds:products] Iniciando geração de %d produtos...", count)

	// Verificar se a tabela products existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'products')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("[seeds:products] Erro ao verificar existência da tabela 'products': %w", err)
	}

	if !exists {
		log.Printf("[seeds:products] Tabela 'products' não existe. Seed de produtos será ignorado.")
		return nil
	}

	// Prepare statement com todos os campos
	stmt, err := db.Prepare(`
        INSERT INTO products 
        (name, detailed_name, description, status, sku, barcode, external_id, 
        coin, price, sales_price, cost_price, stock, type, product_group, 
        product_category, product_subcategory, tags, manufacturer, manufacturer_code, 
        ncm, cest, cnae, origin, created_at, updated_at, images, documents) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
    `)
	if err != nil {
		return fmt.Errorf("[seeds:products] Erro ao preparar inserção de produtos: %w", err)
	}
	defer stmt.Close()

	log.Printf("[seeds:products] Inserção preparada com sucesso.")

	// Valores possíveis para os campos de seleção
	statusOptions := []string{"ativo", "desativado", "descontinuado"}
	coinOptions := []string{"BRL", "USD", "EUR", "CAD", "ADOBE_USD"}
	typeOptions := []string{"produto", "serviço", "assinatura", "licença"}
	groupOptions := []string{"hardware", "software", "serviços", "consumíveis"}
	categoryOptions := []string{"computadores", "periféricos", "redes", "segurança", "marketing", "design"}
	subcategoryOptions := []string{"desktop", "notebook", "servidor", "impressora", "scanner", "roteador"}
	manufacturerOptions := []string{"Dell", "HP", "Lenovo", "Apple", "Samsung", "Asus", "Microsoft", "Adobe"}
	productAdjectives := []string{"Premium", "Profissional", "Avançado", "Básico", "Ultimate", "Lite", "Pro", "Enterprise", "Home", "Business"}
	originOptions := []string{
		string(models.OriginNacionalExceto3_4_5_8),
		string(models.OriginEstrangeiraImportacaoDireta),
		string(models.OriginEstrangeiraMercadoInterno),
		string(models.OriginNacionalConteudoImport40_70),
		string(models.OriginNacionalProcessosProdutivos),
	}

	now := time.Now()

	for i := range count {
		// Gerar preços realistas
		basePrice := gofakeit.Price(100, 10000)
		costPrice := basePrice * 0.65 // Custo é aproximadamente 65% do preço base
		salesPrice := basePrice * 0.9 // Preço promocional é 90% do preço base

		// Gerar tags para o produto
		numTags := gofakeit.Number(1, 5)
		tags := make([]string, numTags)
		for j := 0; j < numTags; j++ {
			tags[j] = gofakeit.ProductCategory()
		}

		// Gerar URLs para imagens
		numImages := gofakeit.Number(1, 4)
		images := make([]string, numImages)
		for j := 0; j < numImages; j++ {
			images[j] = fmt.Sprintf("https://example.com/images/product%d_%d.jpg", i, j)
		}

		// Gerar URLs para documentos
		numDocs := gofakeit.Number(0, 2)
		docs := make([]string, numDocs)
		for j := 0; j < numDocs; j++ {
			docs[j] = fmt.Sprintf("https://example.com/docs/product%d_manual%d.pdf", i, j)
		}

		// Função auxiliar para gerar sequências de dígitos (substitui DigitsN)
		generateDigits := func(n int) string {
			digits := ""
			for i := 0; i < n; i++ {
				digits += fmt.Sprintf("%d", gofakeit.Number(0, 9))
			}
			return digits
		}

		// Selecionar um adjetivo aleatório para o nome detalhado do produto
		randomAdjective := productAdjectives[gofakeit.Number(0, len(productAdjectives)-1)]

		// Gerar dados fictícios para o produto
		product := models.Product{
			Name:               gofakeit.ProductName(),
			DetailedName:       gofakeit.ProductName() + " " + randomAdjective,
			Description:        gofakeit.ProductDescription(),
			Status:             statusOptions[gofakeit.Number(0, len(statusOptions)-1)],
			SKU:                fmt.Sprintf("SKU-%s", generateDigits(8)),
			Barcode:            generateDigits(13),
			ExternalID:         gofakeit.UUID(),
			Coin:               coinOptions[gofakeit.Number(0, len(coinOptions)-1)],
			Price:              basePrice,
			SalesPrice:         salesPrice,
			CostPrice:          costPrice,
			Stock:              gofakeit.Number(0, 1000),
			Type:               typeOptions[gofakeit.Number(0, len(typeOptions)-1)],
			ProductGroup:       groupOptions[gofakeit.Number(0, len(groupOptions)-1)],
			ProductCategory:    categoryOptions[gofakeit.Number(0, len(categoryOptions)-1)],
			ProductSubcategory: subcategoryOptions[gofakeit.Number(0, len(subcategoryOptions)-1)],
			Tags:               pq.StringArray(tags),
			Manufacturer:       manufacturerOptions[gofakeit.Number(0, len(manufacturerOptions)-1)],
			ManufacturerCode:   fmt.Sprintf("MFG-%s", generateDigits(6)),
			NCM:                generateDigits(8),
			CEST:               generateDigits(7),
			CNAE:               generateDigits(7),
			Origin:             originOptions[gofakeit.Number(0, len(originOptions)-1)],
			CreatedAt:          now,
			UpdatedAt:          now,
			Images:             pq.StringArray(images),
			Documents:          pq.StringArray(docs),
		}

		_, err := stmt.Exec(
			product.Name,
			product.DetailedName,
			product.Description,
			product.Status,
			product.SKU,
			product.Barcode,
			product.ExternalID,
			product.Coin,
			product.Price,
			product.SalesPrice,
			product.CostPrice,
			product.Stock,
			product.Type,
			product.ProductGroup,
			product.ProductCategory,
			product.ProductSubcategory,
			pq.Array(product.Tags),
			product.Manufacturer,
			product.ManufacturerCode,
			product.NCM,
			product.CEST,
			product.CNAE,
			product.Origin,
			product.CreatedAt,
			product.UpdatedAt,
			pq.Array(product.Images),
			pq.Array(product.Documents),
		)

		if err != nil {
			return fmt.Errorf("[seeds:products] Erro ao inserir produto #%d: %w", i+1, err)
		}
	}

	log.Printf("[seeds:products] Geração de produtos concluída com sucesso.")
	return nil
}
