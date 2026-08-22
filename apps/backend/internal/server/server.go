package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/lopor-ai/lopor/internal/database"
	"github.com/lopor-ai/lopor/internal/domain/agent"
	"github.com/lopor-ai/lopor/internal/domain/auth"
	"github.com/lopor-ai/lopor/internal/domain/chat"
	"github.com/lopor-ai/lopor/internal/domain/document"
	"github.com/lopor-ai/lopor/internal/domain/job"
	"github.com/lopor-ai/lopor/internal/domain/organization"
	"github.com/lopor-ai/lopor/internal/domain/prompt"
	"github.com/lopor-ai/lopor/internal/domain/rag"
	"github.com/lopor-ai/lopor/internal/domain/workspace"
	"github.com/lopor-ai/lopor/internal/middleware"
	"github.com/lopor-ai/lopor/pkg/ai"
	"github.com/lopor-ai/lopor/pkg/collaboration"
	"github.com/lopor-ai/lopor/pkg/email"
	"github.com/lopor-ai/lopor/pkg/jobqueue"
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
	app.Use(middleware.PrometheusMetrics())
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
	chatRepo := chat.NewRepository(pool)
	chatService := chat.NewService(chatRepo)
	chatHandler := chat.NewHandler(chatService, aiClient)

	ragRepo := rag.NewRepository(cfg.DB.Pool)
	ragService := rag.NewService(ragRepo)
	ragHandler := rag.NewHandler(ragService)

	docRepo := document.NewRepository(cfg.DB.Pool)
	docService := document.NewService(docRepo)
	docHandler := document.NewHandler(docService)

	agentRepo := agent.NewRepository(cfg.DB.Pool)
	agentService := agent.NewService(agentRepo)
	agentHandler := agent.NewHandler(agentService)

	mailer := email.NewMailer("", "", "")
	orgRepo := organization.NewRepository(cfg.DB.Pool)
	orgService := organization.NewService(orgRepo, mailer)
	orgHandler := organization.NewHandler(orgService)

	promptRepo := prompt.NewRepository(cfg.DB.Pool)
	promptService := prompt.NewService(promptRepo)
	promptHandler := prompt.NewHandler(promptService)

	var redisConn *redis.Client
	if cfg.Redis != nil {
		redisConn = cfg.Redis.Client
	}
	queue := jobqueue.NewQueue(redisConn, "lopor_jobs_queue")
	jobHandler := job.NewHandler(queue)

	// API Route Group
	api := app.Group("/api/v1")

	// Auth Endpoints
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Get("/me", middleware.Protected(cfg.JWTSecret), authHandler.GetMe)

	// Async Job Queue Endpoints
	jobsGroup := api.Group("/jobs", middleware.Protected(cfg.JWTSecret))
	jobsGroup.Post("/enqueue", jobHandler.EnqueueJob)
	jobsGroup.Get("/status/:jobId", jobHandler.GetJobStatus)

	// Multi-Model AI Gateway Endpoints
	aiRouter := ai.NewModelRouter()
	api.Get("/ai/models", func(c *fiber.Ctx) error {
		models := aiRouter.GetAvailableModels()
		return response.Success(c, fiber.StatusOK, "Available AI models retrieved", models)
	})

	// Organization & Multi-Tenancy Endpoints
	orgGroup := api.Group("/organizations", middleware.Protected(cfg.JWTSecret))
	orgGroup.Post("/", orgHandler.CreateOrganization)
	orgGroup.Get("/", orgHandler.GetUserOrganizations)
	orgGroup.Get("/:orgId/members", orgHandler.GetOrganizationMembers)
	orgGroup.Post("/:orgId/invite", orgHandler.InviteMember)

	// Workspace Endpoints
	wsGroup := api.Group("/workspaces", middleware.Protected(cfg.JWTSecret))
	wsGroup.Post("/", wsHandler.CreateWorkspace)
	wsGroup.Get("/", wsHandler.GetUserWorkspaces)
	wsGroup.Get("/:id", wsHandler.GetWorkspaceByID)

	// Prompt Templates & Studio Endpoints
	wsGroup.Post("/:wsId/prompts", promptHandler.CreatePrompt)
	wsGroup.Get("/:wsId/prompts", promptHandler.GetWorkspacePrompts)
	wsGroup.Post("/:wsId/prompts/:promptId/execute", promptHandler.SubstituteVariables)
	wsGroup.Delete("/:wsId/prompts/:promptId", promptHandler.DeletePrompt)

	// Chat Endpoints
	wsGroup.Post("/:wsId/chats", chatHandler.CreateChat)
	wsGroup.Get("/:wsId/chats", chatHandler.GetWorkspaceChats)
	wsGroup.Get("/:wsId/chats/:chatId", chatHandler.GetChatDetails)
	wsGroup.Post("/:wsId/chats/:chatId/stream", chatHandler.StreamChatResponse)

	// RAG Vector & File Ingestion Endpoints
	wsGroup.Post("/:wsId/search/semantic", ragHandler.SemanticSearch)
	wsGroup.Post("/:wsId/search/hybrid", ragHandler.HybridSearch)
	wsGroup.Post("/:wsId/ingest", ragHandler.IngestText)
	wsGroup.Post("/:wsId/files/upload", ragHandler.UploadFile)
	wsGroup.Get("/:wsId/files", ragHandler.GetFiles)

	// Documents & Folders Endpoints
	wsGroup.Post("/:wsId/documents", docHandler.CreateDocument)
	wsGroup.Get("/:wsId/documents", docHandler.GetWorkspaceDocuments)
	wsGroup.Get("/:wsId/documents/:docId", docHandler.GetDocumentByID)
	wsGroup.Patch("/:wsId/documents/:docId", docHandler.UpdateDocument)
	wsGroup.Post("/:wsId/folders", docHandler.CreateFolder)
	wsGroup.Get("/:wsId/folders", docHandler.GetWorkspaceFolders)

	// Autonomous AI Agents Endpoints
	wsGroup.Post("/:wsId/agents", agentHandler.CreateAgent)
	wsGroup.Get("/:wsId/agents", agentHandler.GetWorkspaceAgents)
	wsGroup.Post("/:wsId/agents/:agentId/execute", agentHandler.ExecuteAgent)
	wsGroup.Delete("/:wsId/agents/:agentId", agentHandler.DeleteAgent)

	// Real-Time WebSockets Collaborative Editing Route
	app.Use("/ws", collaboration.WebSocketUpgradeMiddleware())
	app.Get("/ws/workspaces/:wsId/documents/:docId", websocket.New(collaboration.HandleWebSocketConnection))

	log.Println("Routes successfully registered in Fiber Core Engine")
	return app
}
