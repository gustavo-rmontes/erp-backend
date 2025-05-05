package main

import (
	"flag"
	"fmt"
	"log"

	"ERP-ONSMART/backend/internal/config"
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/db/seeds"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Flags para controlar seeds
	runSeeds := flag.Bool("seed", false, "Executar seeds para dados de desenvolvimento")
	seedCustomers := flag.Int("customers", 400, "Número de clientes a serem gerados")
	seedProducts := flag.Int("products", 200, "Número de produtos a serem gerados")
	seedOrders := flag.Int("orders", 300, "Número de pedidos a serem gerados")
	seedContacts := flag.Int("contacts", 150, "Número de contatos a serem gerados")
	seedUsers := flag.Int("users", 20, "Número de usuários a serem gerados")
	seedTransactions := flag.Int("transactions", 500, "Número de transações a serem geradas")
	seedCampaigns := flag.Int("campaigns", 30, "Número de campanhas a serem geradas")
	seedRentals := flag.Int("rentals", 100, "Número de aluguéis a serem gerados")
	seedSales := flag.Int("sales", 400, "Número de vendas a serem geradas")
	seedValue := flag.Int64("seed-value", 42, "Valor da seed para reprodutibilidade")
	flag.Parse()

	// Inicializa o logger
	if _, err := logger.InitLogger(); err != nil {
		log.Fatalf("Erro ao inicializar logger: %v", err)
	}

	// Carrega configurações do .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Executa as migrations
	if err := db.RunMigrations(); err != nil {
		// Não aborta a execução em caso de erro nas migrations
		log.Printf("[main.go]: Aviso ao executar migrations: %v", err)
	}

	// Executa seeds se solicitado via flag
	if *runSeeds {
		log.Println("[main.go]: Iniciando geração de dados mock para desenvolvimento...")

		// Obtém conexão com o banco de dados
		database, err := db.OpenDB()
		if err != nil {
			log.Printf("[main.go]: Erro ao conectar ao banco para seeds: %v", err)
		} else {
			// Configura os parâmetros de seed
			seedConfig := seeds.SeedConfig{
				CustomersCount:    *seedCustomers,
				ProductsCount:     *seedProducts,
				OrdersCount:       *seedOrders,
				ContactsCount:     *seedContacts,
				UsersCount:        *seedUsers,
				TransactionsCount: *seedTransactions,
				CampaignsCount:    *seedCampaigns,
				RentalsCount:      *seedRentals,
				SalesCount:        *seedSales,
				Seed:              *seedValue,
			}

			// Executa os seeds
			if err := seeds.ExecuteSeeds(database, seedConfig); err != nil {
				log.Printf("[main.go]: Erro ao executar seeds: %v", err)
			} else {
				log.Println("[main.go]: Seeds executados com sucesso!")
			}
		}
	}

	router := gin.Default()

	// Middleware CORS manual (substitui cors.New)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // ou {"*"} se não usar credenciais
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Configura rotas
	routes.SetupRoutes(router)

	fmt.Printf("Ambiente: %s\n", cfg.Env)
	fmt.Printf("Servidor rodando em http://localhost:%s\n", cfg.Port)

	// Inicia o servidor
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
