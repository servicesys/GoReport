package main

import (
	"log"
	"os"

	"reports-system/internal/app/handlers"
	"reports-system/internal/domain/entities"
	"reports-system/internal/infra/cache"
	"reports-system/internal/infra/database"
	"reports-system/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Inicializar banco de dados
	if os.Getenv("DB_TYPE") == "" || os.Getenv("DB_HOST") == "" {
		log.Fatal("Database connection details are not set in environment variables")
	}

	var db entities.Database
	var err error

	if os.Getenv("DB_TYPE") == "postgres" {
		log.Println("Using Postgres database")
		postgress := database.PostgresDB{}
		db, err = postgress.NewDB()

		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		defer db.Close()

	} else if os.Getenv("DB_TYPE") == "sqlserver" {
		log.Println("Using SQL Server database")
		sqlServer := database.SqlServerDB{}
		db, err = sqlServer.NewDB()
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		defer db.Close()
	} else {
		log.Fatal("Unsupported database type. Please set DB_TYPE to 'postgres' or 'sql	server'")
	}

	// Inicializar cache
	cacheProvider := cache.NewMemoryCache()

	// Inicializar serviços
	confReports := os.Getenv("CONFIG_REPORTS")
	reportService := usecase.NewReportService(db, cacheProvider, confReports) // Exemplo de caminho
	reportHandler := handlers.NewReportHandler(reportService)

	// Configurar Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Swagger
	//app.Get("/swagger/*", fiberSwagger.New(swagger.Config{}))

	// Rotas da API
	api := app.Group("/api/v1")

	// Rotas de relatórios
	api.Get("/reports", reportHandler.GetAvailableReports)
	api.Get("/reports/:report_id", reportHandler.GetReport)
	api.Post("/reports/:report_id", reportHandler.PostReport)

	// Health check
	api.Get("/health", func(c fiber.Ctx) error {
		health := db.Health()
		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": health,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
