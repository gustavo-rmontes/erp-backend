package routes

import (
	accountingHandler "ERP-ONSMART/backend/internal/modules/accounting/handler"
	authHandler "ERP-ONSMART/backend/internal/modules/auth/handler"
	contactHandler "ERP-ONSMART/backend/internal/modules/contact/handler"
	dashboardHandler "ERP-ONSMART/backend/internal/modules/dashboard/handler"
	dropshippingHandler "ERP-ONSMART/backend/internal/modules/dropshipping/handler"
	marketingHandler "ERP-ONSMART/backend/internal/modules/marketing/handler"
	productsHandler "ERP-ONSMART/backend/internal/modules/products/handler"
	rentalHandler "ERP-ONSMART/backend/internal/modules/rental/handler"
	salesHandler "ERP-ONSMART/backend/internal/modules/sales/handler"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configura todas as rotas da aplicação.
func SetupRoutes(router *gin.Engine) {
	// Rota pública de boas-vindas
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Bem-vindo ao ERP Inteligente da On Smart Tech"})
	})

	// Endpoint de teste
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.LoginHandler)
		authGroup.POST("/register", authHandler.RegisterHandler)
		authGroup.GET("/profile", authHandler.ProfileHandler)
		authGroup.DELETE("/:username", authHandler.DeleteUserHandler)
	}

	// Grupo de rotas para o módulo de vendas
	salesGroup := router.Group("/sales")
	{
		salesGroup.GET("/", salesHandler.ListSalesHandler)
		salesGroup.GET("/:id", salesHandler.GetSaleHandler)
		salesGroup.POST("/", salesHandler.CreateSaleHandler)
		salesGroup.PUT("/:id", salesHandler.UpdateSaleHandler)
		salesGroup.DELETE("/:id", salesHandler.DeleteSaleHandler)
	}

	// Grupo de rotas para o módulo de accounting
	accountingGroup := router.Group("/accounting")
	{
		accountingGroup.GET("/", accountingHandler.ListTransactionsHandler)
		accountingGroup.POST("/", accountingHandler.CreateTransactionHandler)
		accountingGroup.PUT("/:id", accountingHandler.UpdateTransactionHandler)
		accountingGroup.DELETE("/:id", accountingHandler.DeleteTransactionHandler)
	}

	// Grupo de rotas para o módulo de marketing
	marketingGroup := router.Group("/marketing")
	{
		marketingGroup.GET("/", marketingHandler.ListCampaignsHandler)
		marketingGroup.POST("/", marketingHandler.CreateCampaignHandler)
		marketingGroup.PUT("/:id", marketingHandler.UpdateCampaignHandler)
		marketingGroup.DELETE("/:id", marketingHandler.DeleteCampaignHandler)
	}

	// Grupo de rotas para o módulo de contatos (clientes e fornecedores)
	contactGroup := router.Group("/contacts")
	{
		contactGroup.GET("/", contactHandler.ListContactsHandler)
		contactGroup.GET("/:id", contactHandler.GetContactByIDHandler)
		contactGroup.POST("/", contactHandler.CreateContactHandler)
		contactGroup.PUT("/:id", contactHandler.UpdateContactHandler)
		contactGroup.DELETE("/:id", contactHandler.DeleteContactHandler)
	}

	//Grupo de rotas para o módulo de produtos
	productGroup := router.Group("/products")
	{
		productGroup.GET("/", productsHandler.ListProductsHandler)
		productGroup.GET("/:id", productsHandler.GetProductByIDHandler)
		productGroup.POST("/", productsHandler.CreateProductHandler)
		productGroup.PUT("/:id", productsHandler.UpdateProductHandler)
		productGroup.DELETE("/:id", productsHandler.DeleteProductHandler)
	}

	//Grupo de rotas para o módulo de locação
	rentalGroup := router.Group("/rentals")
	{
		rentalGroup.GET("/", rentalHandler.ListRentalsHandler)
		rentalGroup.POST("/", rentalHandler.CreateRentalHandler)
		rentalGroup.PUT("/:id", rentalHandler.UpdateRentalHandler)
		rentalGroup.DELETE("/:id", rentalHandler.DeleteRentalHandler)
	}

	//Grupo de rotas para o módulo de garantia
	warrantyGroup := router.Group("/warranties")
	{
		warrantyGroup.GET("/", productsHandler.ListWarrantiesHandler)
		warrantyGroup.POST("/", productsHandler.CreateWarrantyHandler)
		warrantyGroup.PUT("/:id", productsHandler.UpdateWarrantyHandler)
		warrantyGroup.DELETE("/:id", productsHandler.DeleteWarrantyHandler)
	}

	//Grupo de rotas para o módulo de dropshipping
	dropshippingGroup := router.Group("/dropshippings")
	{
		dropshippingGroup.GET("/", dropshippingHandler.ListDropshippingsHandler)
		dropshippingGroup.GET("/:id", dropshippingHandler.GetDropshippingHandler)
		dropshippingGroup.POST("/", dropshippingHandler.CreateDropshippingHandler)
		dropshippingGroup.PUT("/:id", dropshippingHandler.UpdateDropshippingHandler)
		dropshippingGroup.DELETE("/:id", dropshippingHandler.DeleteDropshippingHandler)
	}

	// Dentro de SetupRoutes:
	router.GET("/dashboard", dashboardHandler.DashboardHandler)

}
