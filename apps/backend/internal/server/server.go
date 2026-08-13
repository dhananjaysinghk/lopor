package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lopor-ai/lopor/internal/database"
	"github.com/lopor-ai/lopor/internal/domain/auth"
	"github.com/lopor-ai/lopor/internal/domain/chat"
	"github.com/lopor-ai/lopor/internal/domain/rag"
	"github.com/lopor-ai/lopor/internal/domain/workspace"
	"github.com/lopor-ai/lopor/internal/middleware"
	"github.com/lopor-ai/lopor/pkg/ai"
	"github.com/lopor-ai/lopor/pkg/response"
)

type Config struct {
	Port        string
	JWTSecret   string
	FrontendURL string
	DB          *database.DB
	Redis       *database.RedisClient
}

func NewServer(cfg Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Lopor AI Workspace API",
		ServerHeader: "Lopor-Engine",
	})

	// Middleware Stack
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL + ", http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Healthcheck endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Lopor Engine Core is Healthy", fiber.Map{
			"status": "online",
			"db":     "connected",
		})
	})

	// Setup Repositories & Services
	var pool *pgxpool.Pool
	if cfg.DB != nil {
		pool = cfg.DB.Pool
	}

	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	wsRepo := workspace.NewRepository(pool)
	wsService := workspace.NewService(wsRepo)
	wsHandler := workspace.NewHandler(wsService)

	aiClient := ai.NewClient("", "")
	chatRepo := chat.NewRepository(cfg.DB.Pool)
	chatService := chat.NewService(chatRepo)
	chatHandler := chat.NewHandler(chatService, aiClient)

	ragRepo := rag.NewRepository(cfg.DB.Pool)
	ragService := rag.NewService(ragRepo)
	ragHandler := rag.NewHandler(ragService)

	// API Route Group
	api := app.Group("/api/v1")

	// Auth Endpoints
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Get("/me", middleware.Protected(cfg.JWTSecret), authHandler.GetMe)

	// Workspace Endpoints
	wsGroup := api.Group("/workspaces", middleware.Protected(cfg.JWTSecret))
	wsGroup.Post("/", wsHandler.CreateWorkspace)
	wsGroup.Get("/", wsHandler.GetUserWorkspaces)
	wsGroup.Get("/:id", wsHandler.GetWorkspaceByID)

	// Chat Endpoints
	wsGroup.Post("/:wsId/chats", chatHandler.CreateChat)
	wsGroup.Get("/:wsId/chats", chatHandler.GetWorkspaceChats)
	wsGroup.Get("/:wsId/chats/:chatId", chatHandler.GetChatDetails)
	wsGroup.Post("/:wsId/chats/:chatId/stream", chatHandler.StreamChatResponse)

	// RAG Vector & File Ingestion Endpoints
	wsGroup.Post("/:wsId/search/semantic", ragHandler.SemanticSearch)
	wsGroup.Post("/:wsId/ingest", ragHandler.IngestText)
	wsGroup.Post("/:wsId/files/upload", ragHandler.UploadFile)
	wsGroup.Get("/:wsId/files", ragHandler.GetFiles)

	log.Println("Routes successfully registered in Fiber Core Engine")
	return app
}
