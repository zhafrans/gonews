package app

import (
	"context"
	"gonews/config"
	"gonews/internal/adapter/cloudflare"
	"gonews/internal/adapter/handler"
	"gonews/internal/adapter/repository"
	"gonews/internal/core/service"
	"gonews/lib/auth"
	"gonews/lib/middleware"
	"gonews/lib/pagination"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

func RunServer() {
	cfg := config.NewConfig()
	db,err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatal("Error connecting to database: %v", err)
		return 
	}

	err = os.MkdirAll("./temp/content", 0755)
	if err != nil {
		log.Fatal("Error creating temp directory: %v", err)
		return
	}


	//cloudfare R2

	cdfR2 := cfg.LoadAwsConfig()
	s3Client := s3.NewFromConfig(cdfR2)
	r2Adapter := cloudflare.NewCloudflareAdapter(s3Client, cfg)

	jwt := auth.NewJwt(cfg)
	middlewareAuth := middleware.NewMiddleware(cfg)

	_ = pagination.NewPagination()

	// Repository
	authRepo := repository.NewAuthRepository(db.DB)
	categoryRepo := repository.NewCategoryRepository(db.DB)
	contentRepo := repository.NewContentRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)


	// Service
	authService := service.NewAuthService(authRepo, cfg, jwt)
	categoryService := service.NewCategoryService(categoryRepo)
	contentService := service.NewContentService(contentRepo, cfg, r2Adapter)
	userService := service.NewUserService(userRepo)

	// handler
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	contentHandler := handler.NewContentHandler(contentService)
	userHandler := handler.NewUserHandler(userService)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip} ${status} - ${latency} ${method} ${path}\n",
	}))

	if os.Getenv("APP_ENV") != "production" {

    // 1. Serve file swagger.json ke URL /api/docs/swagger.json
    app.Static("/api/docs", "./docs")

    // 2. Swagger UI diakses via /api/docs
    app.Get("/api/docs/*", swagger.New(swagger.Config{
        URL: "/api/docs/swagger.json", // path yg diload UI
    }))
}

	api := app.Group("/api")
	api.Post("/login", authHandler.Login)

	adminApp := api.Group("/admin")
	adminApp.Use(middlewareAuth.CheckToken())

	// category
	categoryApp := adminApp.Group("/categories")
	categoryApp.Get("/", categoryHandler.GetCategories)
	categoryApp.Post("/", categoryHandler.CreateCategory)
	categoryApp.Put("/:categoryID", categoryHandler.EditCategoryByID)
	categoryApp.Get("/:categoryID", categoryHandler.GetCategoryByID)
	categoryApp.Delete("/:categoryID", categoryHandler.DeleteCategory)

	// content
	contentApp := adminApp.Group("/contents")
	contentApp.Get("/", contentHandler.GetContents)
	contentApp.Post("/", contentHandler.CreateContent)
	contentApp.Put("/:contentID", contentHandler.UpdateContent)
	contentApp.Get("/:contentID", contentHandler.GetContentByID)
	contentApp.Delete("/:contentID", contentHandler.DeleteContent)
	contentApp.Post("/upload-image", contentHandler.UploadImageR2)
	
	// User
	userApp := adminApp.Group("/users")
	userApp.Get("/profile", userHandler.GetUserByID)
	userApp.Put("/update-password", userHandler.UpdatePassword)

	// FE
	feApp := api.Group("/fe")
	feApp.Get("/categories/", categoryHandler.GetCategoryFE)
	feApp.Get("/contents", contentHandler.GetContentWithQuery)
	feApp.Get("/contents/:contentID", contentHandler.GetContentDetail)


	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err := app.Listen(":" + cfg.App.AppPort)
		if err != nil {
			log.Fatal("Error starting server: %v", err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Println("Server shutdown of 5 seconds")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.ShutdownWithContext(ctx)
}