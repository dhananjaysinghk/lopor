package server

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/lopor-ai/lopor/internal/database"
	"github.com/lopor-ai/lopor/internal/domain/agent"
	"github.com/lopor-ai/lopor/internal/domain/audit"
	"github.com/lopor-ai/lopor/internal/domain/auth"
	"github.com/lopor-ai/lopor/internal/domain/chat"
	"github.com/lopor-ai/lopor/internal/domain/document"
	"github.com/lopor-ai/lopor/internal/domain/graph"
	"github.com/lopor-ai/lopor/internal/domain/job"
	"github.com/lopor-ai/lopor/internal/domain/organization"
	"github.com/lopor-ai/lopor/internal/domain/persona"
	"github.com/lopor-ai/lopor/internal/domain/prompt"
	"github.com/lopor-ai/lopor/internal/domain/rag"
	"github.com/lopor-ai/lopor/internal/domain/workspace"
	"github.com/lopor-ai/lopor/internal/middleware"
	"github.com/lopor-ai/lopor/pkg/agenteval"
	"github.com/lopor-ai/lopor/pkg/ai"
	"github.com/lopor-ai/lopor/pkg/anonymizer"
	"github.com/lopor-ai/lopor/pkg/codereview"
	"github.com/lopor-ai/lopor/pkg/codestudio"
	"github.com/lopor-ai/lopor/pkg/collaboration"
	"github.com/lopor-ai/lopor/pkg/contextopt"
	"github.com/lopor-ai/lopor/pkg/diffsynth"
	"github.com/lopor-ai/lopor/pkg/docexport"
	"github.com/lopor-ai/lopor/pkg/docgen"
	"github.com/lopor-ai/lopor/pkg/docvision"
	"github.com/lopor-ai/lopor/pkg/email"
	"github.com/lopor-ai/lopor/pkg/gitops"
	"github.com/lopor-ai/lopor/pkg/jobqueue"
	"github.com/lopor-ai/lopor/pkg/metering"
	"github.com/lopor-ai/lopor/pkg/observability"
	"github.com/lopor-ai/lopor/pkg/prompteval"
	"github.com/lopor-ai/lopor/pkg/response"
	"github.com/lopor-ai/lopor/pkg/sandbox"
	"github.com/lopor-ai/lopor/pkg/search"
	"github.com/lopor-ai/lopor/pkg/secscan"
	"github.com/lopor-ai/lopor/pkg/sqlsynth"
	"github.com/lopor-ai/lopor/pkg/taskplanner"
	"github.com/lopor-ai/lopor/pkg/testgen"
	"github.com/lopor-ai/lopor/pkg/totp"
	"github.com/lopor-ai/lopor/pkg/voice"
	"github.com/lopor-ai/lopor/pkg/webhook"
	"github.com/lopor-ai/lopor/pkg/zipengine"
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

	openAIKey := os.Getenv("OPENAI_API_KEY")
	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	aiClient := ai.NewClient(openAIKey, openAIBaseURL)
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

	personaRepo := persona.NewRepository(cfg.DB.Pool)
	personaService := persona.NewService(personaRepo)
	personaHandler := persona.NewHandler(personaService)

	graphRepo := graph.NewRepository(cfg.DB.Pool)
	graphService := graph.NewService(graphRepo)
	graphHandler := graph.NewHandler(graphService)

	auditRepo := audit.NewRepository(cfg.DB.Pool)
	auditService := audit.NewService(auditRepo)
	auditHandler := audit.NewHandler(auditService)

	var redisConn *redis.Client
	if cfg.Redis != nil {
		redisConn = cfg.Redis.Client
	}
	queue := jobqueue.NewQueue(redisConn, "lopor_jobs_queue")
	jobHandler := job.NewHandler(queue)

	metricsCollector := observability.NewMetricsCollector(pool, redisConn)
	app.Use(func(c *fiber.Ctx) error {
		metricsCollector.IncRequests()
		err := c.Next()
		if c.Response().StatusCode() >= 400 {
			metricsCollector.IncErrors()
		}
		return err
	})

	// API Route Group
	api := app.Group("/api/v1")

	// Auth Endpoints
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Get("/me", middleware.Protected(cfg.JWTSecret), authHandler.GetMe)

	// Enterprise Multi-Factor Authentication (MFA / TOTP) Endpoints
	mfaEngine := totp.NewMFAEngine()
	authGroup.Post("/mfa/setup", middleware.Protected(cfg.JWTSecret), func(c *fiber.Ctx) error {
		userEmail, _ := c.Locals("userEmail").(string)
		if userEmail == "" {
			userEmail = "user@lopor.ai"
		}
		setup, err := mfaEngine.GenerateSetup("Lopor AI Workspace", userEmail)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "MFA_SETUP_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "MFA enrollment setup generated", setup)
	})

	authGroup.Post("/mfa/verify", middleware.Protected(cfg.JWTSecret), func(c *fiber.Ctx) error {
		var req totp.MFAVerificationReq
		if err := c.BodyParser(&req); err != nil || req.Passcode == "" || req.Secret == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Secret and 6-digit passcode are required", nil)
		}
		valid, err := mfaEngine.VerifyCode(req.Secret, req.Passcode)
		if err != nil || !valid {
			return response.Error(c, fiber.StatusUnauthorized, "MFA_INVALID_CODE", "Invalid or expired passcode", nil)
		}
		return response.Success(c, fiber.StatusOK, "MFA verification successful", fiber.Map{"mfa_verified": true})
	})

	authGroup.Post("/mfa/recovery", func(c *fiber.Ctx) error {
		var req totp.MFARecoveryReq
		if err := c.BodyParser(&req); err != nil || req.RecoveryCode == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Recovery code is required", nil)
		}
		return response.Success(c, fiber.StatusOK, "Recovery code accepted", fiber.Map{"recovered": true})
	})

	// Async Job Queue Endpoints
	jobsGroup := api.Group("/jobs", middleware.Protected(cfg.JWTSecret))
	jobsGroup.Post("/enqueue", jobHandler.EnqueueJob)
	jobsGroup.Get("/status/:jobId", jobHandler.GetJobStatus)

	// Admin Metrics & Deep System Health Endpoints
	api.Get("/admin/metrics/summary", middleware.Protected(cfg.JWTSecret), func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "System telemetry metrics summary retrieved", metricsCollector.CollectMetrics())
	})
	api.Get("/admin/health/deep-check", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Deep system health check completed", metricsCollector.PerformDeepHealthCheck(c.Context()))
	})

	// Multi-Model AI Gateway Endpoints
	aiRouter := ai.NewModelRouter()
	api.Get("/ai/models", func(c *fiber.Ctx) error {
		models := aiRouter.GetAvailableModels()
		return response.Success(c, fiber.StatusOK, "Available AI models retrieved", models)
	})

	// Multi-Language Code Execution Sandbox Endpoints
	sandboxCompiler := sandbox.NewCompiler()
	api.Post("/sandbox/execute", middleware.Protected(cfg.JWTSecret), func(c *fiber.Ctx) error {
		type SandboxRequest struct {
			Language string `json:"language"`
			Code     string `json:"code"`
		}

		var req SandboxRequest
		if err := c.BodyParser(&req); err != nil || req.Code == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Language and Code snippet are required", nil)
		}

		result, err := sandboxCompiler.ExecuteCode(c.Context(), sandbox.Language(req.Language), req.Code)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SANDBOX_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Code sandbox execution completed", result)
	})

	// Public Developer API Endpoints
	api.Get("/developer/openapi.json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"openapi": "3.0.3",
			"info": fiber.Map{
				"title":       "Lopor AI Workspace API",
				"version":     "1.0.0",
				"description": "Production REST & SSE streaming endpoints for pgvector RAG, autonomous agents, and document collaboration.",
			},
		})
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

	// Knowledge Graph Endpoints
	wsGroup.Get("/:wsId/graph", graphHandler.GetWorkspaceGraph)

	// Audit Logs Endpoints
	wsGroup.Get("/:wsId/audit-logs", auditHandler.GetWorkspaceAuditLogs)
	wsGroup.Get("/:wsId/audit-logs/export", auditHandler.ExportAuditLogsCSV)

	// License & Usage Metering Endpoints
	meterService := metering.NewMeterService()
	wsGroup.Get("/:wsId/metering/usage", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		stats := meterService.GetWorkspaceUsage(c.Context(), wsID)
		return response.Success(c, fiber.StatusOK, "Workspace usage metering stats retrieved", stats)
	})

	wsGroup.Post("/:wsId/metering/record", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		var req metering.UsageRecordReq
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload", nil)
		}
		updated, err := meterService.RecordUsage(c.Context(), wsID, req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "RECORD_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Usage event recorded successfully", updated)
	})

	wsGroup.Post("/:wsId/metering/tier", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		type TierReq struct {
			Tier string `json:"tier"`
		}
		var req TierReq
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid JSON payload", nil)
		}
		updated := meterService.SetWorkspaceTier(c.Context(), wsID, metering.Tier(req.Tier))
		return response.Success(c, fiber.StatusOK, "Workspace subscription tier updated", updated)
	})

	// Webhook Event Notifications & Dispatcher Endpoints
	webhookDispatcher := webhook.NewDispatcher()
	wsGroup.Post("/:wsId/webhooks", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		type CreateWebhookReq struct {
			TargetURL string              `json:"target_url"`
			Events    []webhook.EventType `json:"events"`
		}
		var req CreateWebhookReq
		if err := c.BodyParser(&req); err != nil || req.TargetURL == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Target URL is required", nil)
		}
		sub, err := webhookDispatcher.RegisterWebhook(wsID, req.TargetURL, req.Events)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "WEBHOOK_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Webhook subscription registered", sub)
	})

	wsGroup.Get("/:wsId/webhooks", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		subs := webhookDispatcher.GetWorkspaceWebhooks(wsID)
		return response.Success(c, fiber.StatusOK, "Workspace webhook subscriptions retrieved", subs)
	})

	wsGroup.Post("/:wsId/webhooks/:webhookId/test", func(c *fiber.Ctx) error {
		webhookID, err := uuid.Parse(c.Params("webhookId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WEBHOOK_ID", "Webhook ID is invalid", nil)
		}
		res, err := webhookDispatcher.DispatchTestEvent(c.Context(), webhookID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "DISPATCH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Test webhook event delivered successfully", res)
	})

	wsGroup.Delete("/:wsId/webhooks/:webhookId", func(c *fiber.Ctx) error {
		webhookID, err := uuid.Parse(c.Params("webhookId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WEBHOOK_ID", "Webhook ID is invalid", nil)
		}
		if err := webhookDispatcher.DeleteWebhook(webhookID); err != nil {
			return response.Error(c, fiber.StatusNotFound, "DELETE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Webhook subscription deleted successfully", nil)
	})

	// Prompt Templates & Studio Endpoints
	promptEvaluator := prompteval.NewEvaluator()
	wsGroup.Post("/:wsId/prompts", promptHandler.CreatePrompt)
	wsGroup.Get("/:wsId/prompts", promptHandler.GetWorkspacePrompts)
	wsGroup.Post("/:wsId/prompts/:promptId/execute", promptHandler.SubstituteVariables)
	wsGroup.Post("/:wsId/prompts/:promptId/benchmark", func(c *fiber.Ctx) error {
		var req prompteval.EvalRequest
		if err := c.BodyParser(&req); err != nil || req.PromptTemplate == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Prompt template is required for benchmark evaluation", nil)
		}
		result, err := promptEvaluator.EvaluatePrompt(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "BENCHMARK_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Prompt benchmark evaluation completed", result)
	})
	wsGroup.Delete("/:wsId/prompts/:promptId", promptHandler.DeletePrompt)

	// AI Personas & System Prompt Management Endpoints
	wsGroup.Post("/:wsId/personas", personaHandler.CreatePersona)
	wsGroup.Get("/:wsId/personas", personaHandler.GetWorkspacePersonas)
	wsGroup.Get("/:wsId/personas/:personaId", personaHandler.GetPersonaByID)
	wsGroup.Put("/:wsId/personas/:personaId", personaHandler.UpdatePersona)
	wsGroup.Post("/:wsId/personas/:personaId/default", personaHandler.SetDefaultPersona)
	wsGroup.Delete("/:wsId/personas/:personaId", personaHandler.DeletePersona)

	// Chat Endpoints
	wsGroup.Post("/:wsId/chats", chatHandler.CreateChat)
	wsGroup.Get("/:wsId/chats", chatHandler.GetWorkspaceChats)
	wsGroup.Get("/:wsId/chats/:chatId", chatHandler.GetChatDetails)
	wsGroup.Post("/:wsId/chats/:chatId/stream", chatHandler.StreamChatResponse)

	// Context Compression & Token Optimization Endpoints
	contextCompressor := contextopt.NewContextCompressor()
	wsGroup.Post("/:wsId/ai/compress-context", func(c *fiber.Ctx) error {
		var req contextopt.CompressionRequest
		if err := c.BodyParser(&req); err != nil || req.Text == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Text content is required for compression", nil)
		}
		result, err := contextCompressor.CompressContext(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "COMPRESSION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Context compressed and token-optimized successfully", result)
	})

	// RAG Vector & File Ingestion Endpoints
	wsGroup.Post("/:wsId/search/semantic", ragHandler.SemanticSearch)
	wsGroup.Post("/:wsId/search/hybrid", ragHandler.HybridSearch)
	wsGroup.Post("/:wsId/ingest", ragHandler.IngestText)
	wsGroup.Post("/:wsId/ingest/url", ragHandler.IngestURL)
	wsGroup.Post("/:wsId/files/upload", ragHandler.UploadFile)
	wsGroup.Get("/:wsId/files", ragHandler.GetFiles)

	// Live Web Grounding Search Endpoints
	webGrounder := search.NewWebGrounder()
	wsGroup.Post("/:wsId/search/web-grounding", func(c *fiber.Ctx) error {
		type SearchRequest struct {
			Query string `json:"query"`
		}
		var req SearchRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_QUERY", "Query text is required", nil)
		}

		res, err := webGrounder.GroundQuery(c.Context(), req.Query)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GROUNDING_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Live web grounding search completed", res)
	})

	// AI Code Studio GitHub Exporter Endpoints
	codeExporter := codestudio.NewExporter()
	wsGroup.Post("/:wsId/code/export-github", func(c *fiber.Ctx) error {
		type ExportReq struct {
			RepoName string                     `json:"repo_name"`
			Scaffold codestudio.ProjectScaffold `json:"scaffold"`
		}
		var req ExportReq
		if err := c.BodyParser(&req); err != nil || req.RepoName == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Repository name is required", nil)
		}

		res, err := codeExporter.ExportToGitHub(c.Context(), req.RepoName, req.Scaffold, "")
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Project scaffold exported to GitHub repository", res)
	})

	// Multi-Format Code & Project Zip Exporter Endpoints
	zipEng := zipengine.NewZipEngine()
	wsGroup.Post("/:wsId/code/export-zip", func(c *fiber.Ctx) error {
		var req zipengine.ZipArchiveRequest
		if err := c.BodyParser(&req); err != nil || len(req.Files) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Files array cannot be empty", nil)
		}

		if req.ArchiveName == "" {
			req.ArchiveName = "lopor_project_export.zip"
		}

		zipBytes, err := zipEng.CreateZipArchive(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "ZIP_EXPORT_FAILED", err.Error(), nil)
		}

		c.Set("Content-Type", "application/zip")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", req.ArchiveName))
		return c.Send(zipBytes)
	})

	// Intelligent Code Refactoring & Patch Diff Synthesizer Endpoints
	diffSynthesizer := diffsynth.NewSynthesizer()
	wsGroup.Post("/:wsId/code/diff-synthesize", func(c *fiber.Ctx) error {
		var req diffsynth.DiffRequest
		if err := c.BodyParser(&req); err != nil || req.OriginalCode == "" || req.ModifiedCode == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Original and modified code snippets are required", nil)
		}

		res, err := diffSynthesizer.SynthesizeDiff(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DIFF_SYNTH_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Unified git diff synthesized successfully", res)
	})

	// Intelligent Code Security Scanner & Secret Leak Detector Endpoints
	codeScanner := secscan.NewScanner()
	wsGroup.Post("/:wsId/code/security-scan", func(c *fiber.Ctx) error {
		var req secscan.ScanRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "File content is required for scanning", nil)
		}

		res, err := codeScanner.ScanCode(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SCAN_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Code security scan completed", res)
	})

	// Enterprise Data Anonymization & PII Redaction Endpoints
	piiAnonymizer := anonymizer.NewPIIAnonymizer()
	wsGroup.Post("/:wsId/security/anonymize", func(c *fiber.Ctx) error {
		var req anonymizer.AnonymizeRequest
		if err := c.BodyParser(&req); err != nil || req.Text == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Text content is required for anonymization", nil)
		}
		res, err := piiAnonymizer.AnonymizeText(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "ANONYMIZE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Data anonymized and PII redacted successfully", res)
	})

	// AI Code Reviewer & Automated Pull Request Critique Endpoints
	codeReviewer := codereview.NewCodeReviewer()
	wsGroup.Post("/:wsId/code/review", func(c *fiber.Ctx) error {
		var req codereview.ReviewRequest
		if err := c.BodyParser(&req); err != nil || req.CodeContent == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Code content is required for review", nil)
		}
		res, err := codeReviewer.ReviewCode(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "REVIEW_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Code review analysis completed", res)
	})

	// AI Automated Unit Test Generator Endpoints
	testGen := testgen.NewGenerator()
	wsGroup.Post("/:wsId/code/test-gen", func(c *fiber.Ctx) error {
		var req testgen.TestGenRequest
		if err := c.BodyParser(&req); err != nil || req.SourceCode == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Source code is required", nil)
		}

		res, err := testGen.GenerateTestSuite(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TEST_GEN_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Automated unit test suite generated successfully", res)
	})

	// Intelligent Markdown Documentation & API Reference Generator Endpoints
	docGenerator := docgen.NewGenerator()
	wsGroup.Post("/:wsId/code/doc-gen", func(c *fiber.Ctx) error {
		var req docgen.DocRequest
		if err := c.BodyParser(&req); err != nil || req.SourceCode == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Source code is required for documentation generation", nil)
		}

		res, err := docGenerator.GenerateDocs(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DOC_GEN_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Technical documentation generated successfully", res)
	})

	// Intelligent SQL Query Generator & Schema Introspection Endpoints
	sqlSynthesizer := sqlsynth.NewSQLSynthesizer()
	wsGroup.Post("/:wsId/sql/generate", func(c *fiber.Ctx) error {
		var req sqlsynth.SQLGenerateRequest
		if err := c.BodyParser(&req); err != nil || req.NaturalQuery == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Natural language query is required", nil)
		}
		res, err := sqlSynthesizer.GenerateSQL(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SQL_GEN_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "SQL query synthesized successfully", res)
	})

	// Intelligent Git Operations & Version Control Endpoints
	gitOpsEngine := gitops.NewGitOpsEngine()
	wsGroup.Post("/:wsId/git/commit", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		var req gitops.CommitRequest
		if err := c.BodyParser(&req); err != nil || len(req.ChangedFiles) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Changed files list is required", nil)
		}
		commit, err := gitOpsEngine.Commit(c.Context(), wsID, req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "COMMIT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Git commit created successfully", commit)
	})

	wsGroup.Get("/:wsId/git/branches", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		branches := gitOpsEngine.GetBranches(wsID)
		return response.Success(c, fiber.StatusOK, "Workspace branches retrieved", branches)
	})

	wsGroup.Post("/:wsId/git/merge-check", func(c *fiber.Ctx) error {
		wsID, err := uuid.Parse(c.Params("wsId"))
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WORKSPACE_ID", "Workspace ID is invalid", nil)
		}
		var req gitops.MergeCheckRequest
		if err := c.BodyParser(&req); err != nil || req.SourceBranch == "" || req.TargetBranch == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Source and target branch names are required", nil)
		}
		res, err := gitOpsEngine.CheckMerge(c.Context(), wsID, req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "MERGE_CHECK_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Git merge check completed", res)
	})

	// Documents & Folders Endpoints
	docConverter := docexport.NewDocumentConverter()
	wsGroup.Post("/:wsId/documents", docHandler.CreateDocument)
	wsGroup.Get("/:wsId/documents", docHandler.GetWorkspaceDocuments)
	wsGroup.Get("/:wsId/documents/:docId", docHandler.GetDocumentByID)
	wsGroup.Post("/:wsId/documents/:id/summarize", docHandler.SummarizeDocument)
	wsGroup.Post("/:wsId/documents/:docId/export", func(c *fiber.Ctx) error {
		var req docexport.ExportRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Document content is required for export", nil)
		}
		res, err := docConverter.ConvertDocument(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "EXPORT_FAILED", err.Error(), nil)
		}
		c.Set("Content-Type", res.ContentType)
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", res.FileName))
		return c.Send(res.Data)
	})
	wsGroup.Patch("/:wsId/documents/:docId", docHandler.UpdateDocument)
	wsGroup.Post("/:wsId/folders", docHandler.CreateFolder)
	wsGroup.Get("/:wsId/folders", docHandler.GetWorkspaceFolders)

	// Autonomous AI Agents Endpoints
	taskPlanner := taskplanner.NewTaskPlanner()
	wsGroup.Post("/:wsId/agents", agentHandler.CreateAgent)
	wsGroup.Get("/:wsId/agents", agentHandler.GetWorkspaceAgents)
	wsGroup.Post("/:wsId/agents/:agentId/execute", agentHandler.ExecuteAgent)
	wsGroup.Post("/:wsId/agents/plan-dag", func(c *fiber.Ctx) error {
		type PlanReq struct {
			Goal  string                 `json:"goal"`
			Nodes []taskplanner.TaskNode `json:"nodes"`
		}
		var req PlanReq
		if err := c.BodyParser(&req); err != nil || len(req.Nodes) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Nodes list cannot be empty", nil)
		}
		plan, err := taskPlanner.PlanDAG(req.Goal, req.Nodes)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "DAG_PLAN_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Autonomous task DAG plan created successfully", plan)
	})
	wsGroup.Post("/:wsId/agents/execute-dag", func(c *fiber.Ctx) error {
		var plan taskplanner.DAGPlan
		if err := c.BodyParser(&plan); err != nil || len(plan.Nodes) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Valid DAG plan is required for execution", nil)
		}
		res, err := taskPlanner.ExecuteDAG(c.Context(), &plan)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DAG_EXEC_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Task DAG executed successfully", res)
	})

	// AI Agent Security Guardrail Verification Endpoints
	guardrailEngine := agenteval.NewGuardrailEngine()
	wsGroup.Post("/:wsId/agents/verify-guardrails", func(c *fiber.Ctx) error {
		var req agenteval.GuardrailRequest
		if err := c.BodyParser(&req); err != nil || (req.PromptText == "" && req.ToolArguments == "") {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Prompt text or tool arguments are required", nil)
		}
		res, err := guardrailEngine.VerifyGuardrails(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GUARDRAIL_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Guardrail verification completed", res)
	})
	wsGroup.Delete("/:wsId/agents/:agentId", agentHandler.DeleteAgent)

	// Real-Time WebSockets Collaborative Editing Route
	app.Use("/ws", collaboration.WebSocketUpgradeMiddleware())
	app.Get("/ws/workspaces/:wsId/documents/:docId", websocket.New(collaboration.HandleWebSocketConnection))

	// Voice Dictation & Audio Transcription Endpoints
	voiceTranscriber := voice.NewTranscriber()
	wsGroup.Post("/:wsId/voice/transcribe", func(c *fiber.Ctx) error {
		fileHeader, err := c.FormFile("audio")
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_AUDIO", "Audio file attachment is required", nil)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "READ_ERROR", "Failed to open audio file", nil)
		}
		defer file.Close()

		buf := make([]byte, fileHeader.Size)
		_, _ = file.Read(buf)

		result, err := voiceTranscriber.TranscribeAudioBytes(c.Context(), buf, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TRANSCRIBE_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Audio transcribed successfully", result)
	})

	// Multi-Modal OCR & Document Vision Extraction Endpoints
	visionExtractor := docvision.NewVisionExtractor()
	wsGroup.Post("/:wsId/vision/extract", func(c *fiber.Ctx) error {
		var req docvision.VisionRequest
		fileHeader, err := c.FormFile("image")
		if err == nil {
			file, err := fileHeader.Open()
			if err == nil {
				defer file.Close()
				buf := make([]byte, fileHeader.Size)
				_, _ = file.Read(buf)
				result, err := visionExtractor.ExtractFromBytes(c.Context(), buf, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), "")
				if err == nil {
					return response.Success(c, fiber.StatusOK, "Document vision extraction completed", result)
				}
			}
		}

		if err := c.BodyParser(&req); err != nil || (req.ImageBase64 == "" && req.FileName == "") {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Image upload or image_base64 is required", nil)
		}

		res, err := visionExtractor.ExtractFromImage(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "VISION_FAILED", err.Error(), nil)
		}

		return response.Success(c, fiber.StatusOK, "Document vision extraction completed", res)
	})

	log.Println("Routes successfully registered in Fiber Core Engine")
	return app
}
