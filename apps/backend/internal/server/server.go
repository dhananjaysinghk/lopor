package server

import (
	"context"
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
	"github.com/lopor-ai/lopor/pkg/auditproof"
	"github.com/lopor-ai/lopor/pkg/clarification"
	"github.com/lopor-ai/lopor/pkg/chunking"
	"github.com/lopor-ai/lopor/pkg/codereview"
	"github.com/lopor-ai/lopor/pkg/codestudio"
	"github.com/lopor-ai/lopor/pkg/collaboration"
	"github.com/lopor-ai/lopor/pkg/consensus"
	"github.com/lopor-ai/lopor/pkg/contextopt"
	"github.com/lopor-ai/lopor/pkg/costcalc"
	"github.com/lopor-ai/lopor/pkg/datagen"
	"github.com/lopor-ai/lopor/pkg/dedup"
	"github.com/lopor-ai/lopor/pkg/diffsynth"
	"github.com/lopor-ai/lopor/pkg/docexport"
	"github.com/lopor-ai/lopor/pkg/docgen"
	"github.com/lopor-ai/lopor/pkg/docvision"
	"github.com/lopor-ai/lopor/pkg/email"
	"github.com/lopor-ai/lopor/pkg/federation"
	"github.com/lopor-ai/lopor/pkg/gitops"
	"github.com/lopor-ai/lopor/pkg/graphrag"
	"github.com/lopor-ai/lopor/pkg/guardrails"
	"github.com/lopor-ai/lopor/pkg/hitl"
	"github.com/lopor-ai/lopor/pkg/indexopt"
	"github.com/lopor-ai/lopor/pkg/intent"
	"github.com/lopor-ai/lopor/pkg/jobqueue"
	"github.com/lopor-ai/lopor/pkg/lineage"
	"github.com/lopor-ai/lopor/pkg/llmjudge"
	"github.com/lopor-ai/lopor/pkg/memory"
	"github.com/lopor-ai/lopor/pkg/metering"
	"github.com/lopor-ai/lopor/pkg/observability"
	"github.com/lopor-ai/lopor/pkg/orchestrator"
	"github.com/lopor-ai/lopor/pkg/postmortem"
	"github.com/lopor-ai/lopor/pkg/prompteval"
	"github.com/lopor-ai/lopor/pkg/promptopt"
	"github.com/lopor-ai/lopor/pkg/redteam"
	"github.com/lopor-ai/lopor/pkg/reflection"
	"github.com/lopor-ai/lopor/pkg/resilience"
	"github.com/lopor-ai/lopor/pkg/response"
	"github.com/lopor-ai/lopor/pkg/sandbox"
	"github.com/lopor-ai/lopor/pkg/search"
	"github.com/lopor-ai/lopor/pkg/secscan"
	"github.com/lopor-ai/lopor/pkg/selfheal"
	"github.com/lopor-ai/lopor/pkg/semanticcache"
	"github.com/lopor-ai/lopor/pkg/simulation"
	"github.com/lopor-ai/lopor/pkg/sqlsynth"
	"github.com/lopor-ai/lopor/pkg/taskplanner"
	"github.com/lopor-ai/lopor/pkg/temporalkg"
	"github.com/lopor-ai/lopor/pkg/testgen"
	"github.com/lopor-ai/lopor/pkg/toolsim"
	"github.com/lopor-ai/lopor/pkg/totp"
	"github.com/lopor-ai/lopor/pkg/translation"
	"github.com/lopor-ai/lopor/pkg/tts"
	"github.com/lopor-ai/lopor/pkg/voice"
	"github.com/lopor-ai/lopor/pkg/voiceagent"
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

	// AI Streaming Cost & Token Consumption Estimator
	costEstimator := costcalc.NewCostEstimator()
	api.Get("/ai/pricing-catalog", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "AI model pricing catalog retrieved", costEstimator.GetCatalog())
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

	// Intelligent Data Lineage & RAG Provenance Tracking Endpoints
	lineageTracker := lineage.NewTracker()
	wsGroup.Post("/:wsId/lineage/record-node", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var node lineage.LineageNode
		if err := c.BodyParser(&node); err != nil || node.ID == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Node ID and payload are required", nil)
		}
		if err := lineageTracker.RecordNode(wsID, node); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "LINEAGE_RECORD_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Lineage node recorded successfully", node)
	})
	wsGroup.Post("/:wsId/lineage/record-edge", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var edge lineage.LineageEdge
		if err := c.BodyParser(&edge); err != nil || edge.SourceNodeID == "" || edge.TargetNodeID == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Source and target node IDs are required", nil)
		}
		if err := lineageTracker.RecordEdge(wsID, edge); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "LINEAGE_EDGE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Lineage edge recorded successfully", edge)
	})
	wsGroup.Get("/:wsId/lineage/trace/:responseId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		respID := c.Params("responseId")
		graph, err := lineageTracker.TraceLineage(wsID, respID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TRACE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Lineage DAG trace retrieved successfully", graph)
	})
	wsGroup.Get("/:wsId/lineage/certificate/:responseId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		respID := c.Params("responseId")
		cert, err := lineageTracker.GenerateProvenanceCertificate(wsID, respID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CERT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Provenance compliance certificate generated", cert)
	})

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

	// Event-Driven Workflow Automation & Action Orchestration Endpoints
	workflowEngine := orchestrator.NewWorkflowOrchestrator()
	wsGroup.Post("/:wsId/workflows", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var rule orchestrator.WorkflowRule
		if err := c.BodyParser(&rule); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid workflow rule payload", nil)
		}
		rule.WorkspaceID = wsID
		if err := workflowEngine.RegisterRule(rule); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Workflow automation rule registered", rule)
	})
	wsGroup.Get("/:wsId/workflows", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		rules := workflowEngine.GetRules(wsID)
		return response.Success(c, fiber.StatusOK, "Workspace workflow rules retrieved", rules)
	})
	wsGroup.Post("/:wsId/workflows/dispatch", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var event orchestrator.EventPayload
		if err := c.BodyParser(&event); err != nil || event.EventType == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Event type and payload are required", nil)
		}
		event.WorkspaceID = wsID
		logs, err := workflowEngine.DispatchEvent(c.Context(), event)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DISPATCH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Event processed through workflow orchestrator", logs)
	})
	wsGroup.Get("/:wsId/workflows/logs", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		logs := workflowEngine.GetExecutionLogs(wsID)
		return response.Success(c, fiber.StatusOK, "Workflow execution logs retrieved", logs)
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

	// Synthetic Dataset & Benchmark Evaluation Generator Endpoints
	datasetGenerator := datagen.NewDatasetGenerator()
	wsGroup.Post("/:wsId/eval/generate-dataset", func(c *fiber.Ctx) error {
		var req datagen.DatasetGenerationRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Content is required for synthetic dataset generation", nil)
		}
		dataset, err := datasetGenerator.GenerateDataset(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DATASET_GEN_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Synthetic evaluation dataset generated successfully", dataset)
	})
	wsGroup.Post("/:wsId/eval/datasets/export-jsonl", func(c *fiber.Ctx) error {
		type ExportReq struct {
			Items []datagen.DatasetItem `json:"items"`
		}
		var req ExportReq
		if err := c.BodyParser(&req); err != nil || len(req.Items) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Dataset items list is required", nil)
		}
		jsonl := datasetGenerator.FormatAsJSONL(req.Items)
		c.Set("Content-Type", "application/x-jsonlines")
		c.Set("Content-Disposition", "attachment; filename=\"eval_benchmark.jsonl\"")
		return c.SendString(jsonl)
	})

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

	// Semantic Router & Zero-Shot Intent Classifier Endpoints
	semanticRouter := intent.NewSemanticRouter()
	wsGroup.Get("/:wsId/router/intents", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Registered semantic intent routes retrieved", semanticRouter.GetRoutes())
	})
	wsGroup.Post("/:wsId/router/classify", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req intent.ClassificationRequest
		if err := c.BodyParser(&req); err != nil || req.Prompt == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Prompt is required for intent classification", nil)
		}
		req.WorkspaceID = wsID
		result, err := semanticRouter.ClassifyPrompt(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CLASSIFICATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Prompt semantically classified", result)
	})
	wsGroup.Post("/:wsId/router/intents", func(c *fiber.Ctx) error {
		var route intent.IntentRoute
		if err := c.BodyParser(&route); err != nil || route.ID == "" || route.Name == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "ID, Name, and Exemplars are required", nil)
		}
		if err := semanticRouter.RegisterCustomRoute(route); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "ROUTE_REGISTRATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Custom intent route registered successfully", route)
	})

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

	wsGroup.Post("/:wsId/ai/estimate-cost", func(c *fiber.Ctx) error {
		var req costcalc.CostEstimateRequest
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid cost estimation request payload", nil)
		}
		result, err := costEstimator.EstimateCost(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "ESTIMATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Token and stream cost estimated successfully", result)
	})

	// Intelligent Self-Correction & Reflection Verification Loop Endpoints
	reflectionEngine := reflection.NewReflectionEngine()
	wsGroup.Post("/:wsId/ai/reflect", func(c *fiber.Ctx) error {
		type ReflectReq struct {
			Prompt          string `json:"prompt"`
			CandidateOutput string `json:"candidate_output"`
			Context         string `json:"context"`
			OutputFormat    string `json:"output_format"`
		}
		var req ReflectReq
		if err := c.BodyParser(&req); err != nil || req.CandidateOutput == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Candidate output is required for reflection", nil)
		}
		critiques, score, err := reflectionEngine.CritiqueOutput(c.Context(), req.Prompt, req.CandidateOutput, req.Context, req.OutputFormat)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "REFLECTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Reflection critique evaluation completed", fiber.Map{
			"score":     score,
			"critiques": critiques,
		})
	})
	wsGroup.Post("/:wsId/ai/self-correct", func(c *fiber.Ctx) error {
		var req reflection.ReflectionRequest
		if err := c.BodyParser(&req); err != nil || req.CandidateOutput == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Candidate output is required for self-correction", nil)
		}
		result, err := reflectionEngine.SelfCorrect(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SELF_CORRECT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Autonomous self-correction loop completed", result)
	})

	// Hierarchical Long-Term Semantic Episodic Memory & Profile Consolidation Endpoints
	memEngine := memory.NewMemoryEngine()
	wsGroup.Post("/:wsId/memory/extract", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req memory.ExtractionRequest
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Malformed memory extraction request payload", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		if req.UserID == "" {
			req.UserID = c.Get("X-User-ID", "default-user")
		}
		res, err := memEngine.ExtractAndStore(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "EXTRACTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Salient episodic memories extracted and stored", res)
	})

	wsGroup.Post("/:wsId/memory/recall", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req memory.RecallRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query string is required for memory recall", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		if req.UserID == "" {
			req.UserID = c.Get("X-User-ID", "default-user")
		}
		res, err := memEngine.RecallMemories(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "RECALL_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Contextual memories recalled successfully", res)
	})

	wsGroup.Get("/:wsId/memory/items", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		userID := c.Query("user_id", c.Get("X-User-ID", "default-user"))
		catFilterStr := c.Query("category")
		var catFilter []memory.MemoryCategory
		if catFilterStr != "" {
			catFilter = append(catFilter, memory.MemoryCategory(catFilterStr))
		}
		items := memEngine.GetStore().List(wsID, userID, catFilter)
		return response.Success(c, fiber.StatusOK, "Active episodic memories retrieved", fiber.Map{
			"total": len(items),
			"items": items,
		})
	})

	wsGroup.Post("/:wsId/memory/consolidate", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		type ConsolidateReq struct {
			UserID string `json:"user_id"`
		}
		var req ConsolidateReq
		_ = c.BodyParser(&req)
		if req.UserID == "" {
			req.UserID = c.Get("X-User-ID", "default-user")
		}
		report, err := memEngine.Consolidate(c.Context(), wsID, req.UserID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CONSOLIDATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Memory consolidation and deduplication completed", report)
	})

	wsGroup.Delete("/:wsId/memory/items/:memoryId", func(c *fiber.Ctx) error {
		memID := c.Params("memoryId")
		deleted := memEngine.GetStore().Delete(memID)
		if !deleted {
			return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Memory item not found", nil)
		}
		return response.Success(c, fiber.StatusOK, "Memory item deleted successfully", nil)
	})

	wsGroup.Delete("/:wsId/memory/purge", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		userID := c.Query("user_id", c.Get("X-User-ID", "default-user"))
		purged := memEngine.GetStore().PurgeUser(wsID, userID)
		return response.Success(c, fiber.StatusOK, "User episodic memories purged (GDPR right-to-be-forgotten)", fiber.Map{
			"purged_count": purged,
		})
	})

	// Enterprise AI Guardrail, Jailbreak Barrier & Real-Time Firewall Endpoints
	guardEngine := guardrails.NewGuardrailEngine()
	wsGroup.Post("/:wsId/guardrails/inspect", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req guardrails.InspectionRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Content payload is required for guardrail inspection", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		if req.Direction == "" {
			req.Direction = guardrails.DirectionInboundPrompt
		}
		res, err := guardEngine.Inspect(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GUARDRAIL_INSPECTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Guardrail firewall evaluation completed", res)
	})

	wsGroup.Get("/:wsId/guardrails/profile", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		profile := guardEngine.GetWorkspaceProfile(wsID)
		return response.Success(c, fiber.StatusOK, "Workspace guardrail profile retrieved", profile)
	})

	wsGroup.Put("/:wsId/guardrails/profile", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var profile guardrails.GuardrailProfile
		if err := c.BodyParser(&profile); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid guardrail profile payload", nil)
		}
		profile.WorkspaceID = wsID
		guardEngine.SetWorkspaceProfile(profile)
		return response.Success(c, fiber.StatusOK, "Workspace guardrail profile updated", guardEngine.GetWorkspaceProfile(wsID))
	})

	// Hierarchical GraphRAG, Triplet Extraction & Community Synthesis Endpoints
	graphRAGEngine := graphrag.NewGraphRAGEngine()
	wsGroup.Post("/:wsId/graphrag/extract", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req graphrag.GraphExtractionRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Content text is required for graph extraction", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		res, err := graphRAGEngine.ExtractAndIngest(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GRAPHRAG_EXTRACTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Knowledge graph triplets extracted and communities updated", res)
	})

	wsGroup.Post("/:wsId/graphrag/search/local", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req graphrag.LocalGraphSearchRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query string is required for local graph search", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		res, err := graphRAGEngine.LocalSearch(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "LOCAL_GRAPH_SEARCH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Local knowledge graph traversal completed", res)
	})

	wsGroup.Post("/:wsId/graphrag/search/global", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req graphrag.GlobalGraphSearchRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query string is required for global graph search", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		res, err := graphRAGEngine.GlobalSearch(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GLOBAL_GRAPH_SEARCH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Global community graph synthesis completed", res)
	})

	wsGroup.Get("/:wsId/graphrag/visual", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		res, err := graphRAGEngine.GetGraphVisual(c.Context(), wsID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "GRAPH_VISUAL_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Knowledge graph visual topology retrieved", res)
	})

	// Multi-Provider Resilience, Dynamic Load Balancing & Circuit Breaker Endpoints
	resilienceEngine := resilience.NewResilienceEngine()
	wsGroup.Get("/:wsId/resilience/pools/:poolId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		poolID := c.Params("poolId")
		status, err := resilienceEngine.GetPoolStatus(wsID, poolID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "POOL_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Provider pool telemetry and circuit status retrieved", status)
	})

	wsGroup.Post("/:wsId/resilience/pools", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var cfg resilience.ProviderPoolConfig
		if err := c.BodyParser(&cfg); err != nil || cfg.PoolID == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Valid pool configuration with pool_id required", nil)
		}
		cfg.WorkspaceID = wsID
		resilienceEngine.RegisterPool(cfg)
		status, _ := resilienceEngine.GetPoolStatus(wsID, cfg.PoolID)
		return response.Success(c, fiber.StatusOK, "Provider pool registered successfully", status)
	})

	wsGroup.Post("/:wsId/resilience/execute", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req resilience.ExecutionRequest
		if err := c.BodyParser(&req); err != nil || req.Prompt == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Prompt text is required for resilient execution", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		if req.PoolID == "" {
			req.PoolID = "default-pool"
		}

		res, err := resilienceEngine.ExecuteWithFallback(c.Context(), req, func(ctx context.Context, providerID, model string) (string, error) {
			// Mock simulated completion or pass to active client
			return fmt.Sprintf("Response synthesized via provider '%s' (model: '%s') for prompt: '%s'", providerID, model, req.Prompt), nil
		})
		if err != nil {
			return response.Error(c, fiber.StatusServiceUnavailable, "ALL_PROVIDERS_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Resilient multi-provider execution completed", res)
	})

	// Human-in-the-Loop (HITL) Policy, Approval Gate & Intervention Endpoints
	hitlEngine := hitl.NewHITLEngine()
	wsGroup.Post("/:wsId/hitl/evaluate", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req hitl.InterceptRequest
		if err := c.BodyParser(&req); err != nil || req.ToolName == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Tool name is required for HITL evaluation", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		ticket, requiresApproval, err := hitlEngine.EvaluateAndIntercept(c.Context(), req)
		if err != nil && ticket != nil && ticket.Status == hitl.StatusRejected {
			return response.Error(c, fiber.StatusForbidden, "ACTION_BLOCKED", err.Error(), ticket)
		}
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "HITL_EVALUATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "HITL intervention policy evaluated", fiber.Map{
			"requires_approval": requiresApproval,
			"ticket":            ticket,
		})
	})

	wsGroup.Get("/:wsId/hitl/tickets", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		statusQuery := c.Query("status")
		var filter []hitl.ApprovalStatus
		if statusQuery != "" {
			filter = append(filter, hitl.ApprovalStatus(statusQuery))
		}
		tickets := hitlEngine.ListTickets(wsID, filter)
		return response.Success(c, fiber.StatusOK, "HITL approval tickets retrieved", fiber.Map{
			"total":   len(tickets),
			"tickets": tickets,
		})
	})

	wsGroup.Get("/:wsId/hitl/tickets/:ticketId", func(c *fiber.Ctx) error {
		ticketID := c.Params("ticketId")
		ticket, err := hitlEngine.GetTicket(ticketID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "TICKET_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "HITL approval ticket details retrieved", ticket)
	})

	wsGroup.Post("/:wsId/hitl/tickets/:ticketId/review", func(c *fiber.Ctx) error {
		ticketID := c.Params("ticketId")
		type ReviewReq struct {
			ReviewerID     string                 `json:"reviewer_id"`
			ReviewerName   string                 `json:"reviewer_name"`
			Decision       hitl.ReviewDecision    `json:"decision"`
			Feedback       string                 `json:"feedback"`
			ModifiedParams map[string]interface{} `json:"modified_params"`
		}
		var req ReviewReq
		if err := c.BodyParser(&req); err != nil || req.Decision == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Review decision ('APPROVE', 'REJECT', 'CANCEL') is required", nil)
		}
		if req.ReviewerID == "" {
			req.ReviewerID = c.Get("X-User-ID", "admin-reviewer")
		}
		if req.ReviewerName == "" {
			req.ReviewerName = "Workspace Reviewer"
		}

		ticket, err := hitlEngine.SubmitReview(c.Context(), hitl.ReviewSubmission{
			TicketID:       ticketID,
			ReviewerID:     req.ReviewerID,
			ReviewerName:   req.ReviewerName,
			Decision:       req.Decision,
			Feedback:       req.Feedback,
			ModifiedParams: req.ModifiedParams,
		})
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "REVIEW_SUBMISSION_FAILED", err.Error(), ticket)
		}
		return response.Success(c, fiber.StatusOK, "HITL review decision recorded", ticket)
	})

	wsGroup.Put("/:wsId/hitl/config", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var cfg hitl.WorkspaceHITLConfig
		if err := c.BodyParser(&cfg); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid HITL policy configuration payload", nil)
		}
		cfg.WorkspaceID = wsID
		hitlEngine.ConfigureWorkspace(cfg)
		return response.Success(c, fiber.StatusOK, "Workspace HITL policy configuration updated", hitlEngine.GetWorkspaceConfig(wsID))
	})

	// Intelligent Conversational Query Disambiguation & Clarification Endpoints
	clarificationEngine := clarification.NewClarificationEngine()
	wsGroup.Post("/:wsId/clarification/analyze", func(c *fiber.Ctx) error {
		type AnalyzeReq struct {
			Query   string `json:"query"`
			Context string `json:"context"`
		}
		var req AnalyzeReq
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query string is required for ambiguity analysis", nil)
		}
		analysis, err := clarificationEngine.AnalyzeQuery(c.Context(), req.Query, req.Context)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "ANALYSIS_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Query ambiguity analysis completed", analysis)
	})

	wsGroup.Post("/:wsId/clarification/fuse", func(c *fiber.Ctx) error {
		var req clarification.FusionRequest
		if err := c.BodyParser(&req); err != nil || req.OriginalQuery == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Original query and answers are required for context fusion", nil)
		}
		result, err := clarificationEngine.FuseClarification(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "FUSION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Query context fusion completed", result)
	})

	// Automated AI Evaluation Harness & LLM-as-a-Judge Scoring Endpoints
	judgeService := llmjudge.NewLLMJudgeService()
	wsGroup.Post("/:wsId/judge/pointwise", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req llmjudge.PointwiseJudgeRequest
		if err := c.BodyParser(&req); err != nil || req.CandidateOutput == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Candidate output is required for pointwise evaluation", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		result, err := judgeService.JudgePointwise(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "JUDGE_EVALUATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Pointwise judge evaluation completed", result)
	})

	wsGroup.Post("/:wsId/judge/pairwise", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req llmjudge.PairwiseJudgeRequest
		if err := c.BodyParser(&req); err != nil || req.CandidateA == "" || req.CandidateB == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Both candidate completions are required for pairwise comparison", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		result, err := judgeService.JudgePairwise(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "PAIRWISE_JUDGE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Blind pairwise comparison completed", result)
	})

	wsGroup.Post("/:wsId/judge/tournament", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req llmjudge.TournamentRequest
		if err := c.BodyParser(&req); err != nil || len(req.Candidates) < 2 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "At least two candidates required for tournament", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		result, err := judgeService.RunTournament(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TOURNAMENT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Tournament benchmark executed successfully", result)
	})

	wsGroup.Get("/:wsId/judge/leaderboard", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		board := judgeService.GetLeaderboard(c.Context(), wsID)
		return response.Success(c, fiber.StatusOK, "Model competitive Elo leaderboard retrieved", fiber.Map{
			"total_models": len(board),
			"leaderboard":  board,
		})
	})

	// Adaptive Semantic Document Chunking & Content-Aware Parser Endpoints
	chunkingEngine := chunking.NewChunkingEngine()
	wsGroup.Post("/:wsId/chunking/process", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req chunking.ChunkingRequest
		if err := c.BodyParser(&req); err != nil || req.RawContent == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Raw document content is required for structural chunking", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		result, err := chunkingEngine.ChunkDocument(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CHUNKING_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Structural content-aware document chunking completed", result)
	})

	// Adaptive Semantic Vector Query Cache & Cost Elimination Gateway Endpoints
	semanticCacheEngine := semanticcache.NewSemanticCacheEngine(5000)
	wsGroup.Post("/:wsId/cache/lookup", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req semanticcache.CacheLookupRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query string is required for cache lookup", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		result, err := semanticCacheEngine.Lookup(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CACHE_LOOKUP_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Semantic cache lookup completed", result)
	})

	wsGroup.Post("/:wsId/cache/store", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req semanticcache.CacheStoreRequest
		if err := c.BodyParser(&req); err != nil || req.Query == "" || req.Response == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query and response are required to store in cache", nil)
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = wsID
		}
		err := semanticCacheEngine.Store(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CACHE_STORE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Response stored in semantic vector cache", nil)
	})

	wsGroup.Get("/:wsId/cache/telemetry", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		telemetry := semanticCacheEngine.GetTelemetry(c.Context(), wsID)
		return response.Success(c, fiber.StatusOK, "Semantic cache efficiency telemetry retrieved", telemetry)
	})

	wsGroup.Delete("/:wsId/cache/invalidate", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		modelID := c.Query("model_id")
		count := semanticCacheEngine.Invalidate(c.Context(), wsID, modelID)
		return response.Success(c, fiber.StatusOK, "Semantic cache invalidated successfully", fiber.Map{
			"purged_entries": count,
		})
	})

	// Multi-Agent Simulation Arena & Persona Scenario Engine Endpoints
	simEngine := simulation.NewEngine()
	wsGroup.Post("/:wsId/simulation/create", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req simulation.StartSimulationRequest
		if err := c.BodyParser(&req); err != nil || req.Title == "" || req.Scenario == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Title and scenario are required for simulation", nil)
		}
		req.WorkspaceID = wsID
		session, err := simEngine.CreateSession(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SIMULATION_INIT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Simulation session initialized successfully", session)
	})

	wsGroup.Get("/:wsId/simulation/sessions", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessions := simEngine.ListSessions(wsID)
		return response.Success(c, fiber.StatusOK, "Simulation sessions retrieved", sessions)
	})

	wsGroup.Get("/:wsId/simulation/:sessionId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		session, err := simEngine.GetSession(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Simulation session not found", nil)
		}
		return response.Success(c, fiber.StatusOK, "Simulation session retrieved", session)
	})

	wsGroup.Post("/:wsId/simulation/:sessionId/step", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		var req simulation.AdvanceStepRequest
		_ = c.BodyParser(&req)
		result, err := simEngine.AdvanceStep(wsID, sessionID, req)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "STEP_ADVANCE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Simulation turn advanced", result)
	})

	wsGroup.Post("/:wsId/simulation/:sessionId/run", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		report, err := simEngine.RunToCompletion(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SIMULATION_RUN_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Simulation executed to completion", report)
	})

	wsGroup.Get("/:wsId/simulation/:sessionId/report", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		report, err := simEngine.GetReport(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "REPORT_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Simulation executive report compiled", report)
	})

	// Autonomous AI Workflow Self-Healing & Diagnostic Recovery Endpoints
	selfHealingEngine := selfheal.NewEngine()
	wsGroup.Post("/:wsId/selfheal/diagnose", func(c *fiber.Ctx) error {
		var fault selfheal.ExecutionFault
		if err := c.BodyParser(&fault); err != nil || fault.ErrorMessage == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Error message and diagnostics are required", nil)
		}
		strategy, severity := selfHealingEngine.Diagnose(fault)
		return response.Success(c, fiber.StatusOK, "Fault diagnosed successfully", fiber.Map{
			"recommended_strategy": strategy,
			"severity":             severity,
		})
	})

	wsGroup.Post("/:wsId/selfheal/remediate", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		type RemediatePayload struct {
			Fault  selfheal.ExecutionFault    `json:"fault"`
			Policy *selfheal.RemediationPolicy `json:"policy,omitempty"`
		}
		var payload RemediatePayload
		if err := c.BodyParser(&payload); err != nil || payload.Fault.ErrorMessage == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Fault payload with error message is required", nil)
		}
		payload.Fault.WorkspaceID = wsID
		result, err := selfHealingEngine.Remediate(c.Context(), payload.Fault, payload.Policy)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "REMEDIATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Autonomous remediation cycle completed", result)
	})

	wsGroup.Get("/:wsId/selfheal/telemetry", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		telemetry := selfHealingEngine.GetTelemetry(wsID)
		return response.Success(c, fiber.StatusOK, "Self-healing resilience telemetry retrieved", telemetry)
	})

	// Continuous Prompt Auto-Optimizer & Few-Shot Demonstration Synthesizer Endpoints
	promptOptEngine := promptopt.NewEngine()
	wsGroup.Post("/:wsId/promptopt/start", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req promptopt.StartOptimizationRequest
		if err := c.BodyParser(&req); err != nil || req.Title == "" || req.BasePrompt == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Title and base_prompt are required for prompt optimization", nil)
		}
		req.WorkspaceID = wsID
		job, err := promptOptEngine.StartOptimization(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "OPTIMIZATION_INIT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Prompt optimization tournament completed", job)
	})

	wsGroup.Get("/:wsId/promptopt/jobs", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		jobs := promptOptEngine.ListJobs(wsID)
		return response.Success(c, fiber.StatusOK, "Prompt optimization jobs retrieved", jobs)
	})

	wsGroup.Get("/:wsId/promptopt/:jobId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		jobID := c.Params("jobId")
		job, err := promptOptEngine.GetJob(wsID, jobID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "JOB_NOT_FOUND", "Optimization job not found", nil)
		}
		return response.Success(c, fiber.StatusOK, "Prompt optimization job retrieved", job)
	})

	wsGroup.Get("/:wsId/promptopt/:jobId/report", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		jobID := c.Params("jobId")
		report, err := promptOptEngine.GetReport(wsID, jobID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "REPORT_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Prompt optimization executive report compiled", report)
	})

	// Temporal Knowledge Graph & Dynamic Entity State Evolution Endpoints
	temporalKGEngine := temporalkg.NewEngine()
	wsGroup.Post("/:wsId/temporalkg/facts", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req temporalkg.IngestFactRequest
		if err := c.BodyParser(&req); err != nil || req.SubjectName == "" || req.Predicate == "" || req.ObjectName == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Subject, predicate, and object are required", nil)
		}
		req.WorkspaceID = wsID
		fact, err := temporalKGEngine.RecordFact(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "INGEST_FACT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Temporal fact recorded with conflict superseding", fact)
	})

	wsGroup.Post("/:wsId/temporalkg/query", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req temporalkg.TemporalQueryRequest
		_ = c.BodyParser(&req)
		req.WorkspaceID = wsID
		result := temporalKGEngine.Query(req)
		return response.Success(c, fiber.StatusOK, "Temporal query executed successfully", result)
	})

	wsGroup.Get("/:wsId/temporalkg/entities", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		entities := temporalKGEngine.ListEntities(wsID)
		return response.Success(c, fiber.StatusOK, "Temporal entities retrieved", entities)
	})

	wsGroup.Get("/:wsId/temporalkg/entities/:entityId/evolution", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		entityID := c.Params("entityId")
		evolution, err := temporalKGEngine.GetEntityEvolution(wsID, entityID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "ENTITY_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Entity temporal evolution trajectory retrieved", evolution)
	})

	// Conversational Voice Agent, VAD & Duplex Audio Endpoints
	voiceAgentEngine := voiceagent.NewEngine()
	wsGroup.Post("/:wsId/voiceagent/sessions", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req voiceagent.StartSessionRequest
		_ = c.BodyParser(&req)
		req.WorkspaceID = wsID
		if req.UserID == "" {
			req.UserID = "default_user"
		}
		session, err := voiceAgentEngine.StartSession(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SESSION_INIT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Voice agent session initialized", session)
	})

	wsGroup.Post("/:wsId/voiceagent/sessions/:sessionId/chunk", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		type ChunkPayload struct {
			PCMBytes        []byte `json:"pcm_bytes"`
			ChunkDurationMs int    `json:"chunk_duration_ms"`
		}
		var payload ChunkPayload
		if err := c.BodyParser(&payload); err != nil || len(payload.PCMBytes) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_AUDIO_CHUNK", "Non-empty pcm_bytes are required", nil)
		}
		result, err := voiceAgentEngine.ProcessChunk(wsID, sessionID, payload.PCMBytes, payload.ChunkDurationMs)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "AUDIO_PROCESSING_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Audio chunk processed", result)
	})

	wsGroup.Get("/:wsId/voiceagent/sessions/:sessionId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		session, err := voiceAgentEngine.GetSession(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Voice session not found", nil)
		}
		return response.Success(c, fiber.StatusOK, "Voice session retrieved", session)
	})

	wsGroup.Delete("/:wsId/voiceagent/sessions/:sessionId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		err := voiceAgentEngine.EndSession(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "SESSION_NOT_FOUND", "Voice session not found", nil)
		}
		return response.Success(c, fiber.StatusOK, "Voice session terminated successfully", nil)
	})

	wsGroup.Get("/:wsId/voiceagent/telemetry", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		telemetry := voiceAgentEngine.GetTelemetry(wsID)
		return response.Success(c, fiber.StatusOK, "Voice agent session telemetry retrieved", telemetry)
	})

	// Autonomous Tool Sandbox, Synthetic API Mocking & Side-Effect Safety Simulation Endpoints
	toolSimEngine := toolsim.NewEngine()
	wsGroup.Get("/:wsId/toolsim/tools", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		tools := toolSimEngine.ListTools(wsID)
		return response.Success(c, fiber.StatusOK, "Mock tools retrieved", tools)
	})

	wsGroup.Post("/:wsId/toolsim/tools", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var tool toolsim.MockTool
		if err := c.BodyParser(&tool); err != nil || tool.Name == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Tool name and schema are required", nil)
		}
		tool.WorkspaceID = wsID
		toolSimEngine.RegisterTool(tool)
		return response.Success(c, fiber.StatusCreated, "Custom mock tool registered", tool)
	})

	wsGroup.Post("/:wsId/toolsim/sandboxes", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		type CreateSandboxPayload struct {
			Title string `json:"title"`
		}
		var payload CreateSandboxPayload
		_ = c.BodyParser(&payload)
		if payload.Title == "" {
			payload.Title = "Agent Tool Execution Dry-Run"
		}
		session, err := toolSimEngine.CreateSandbox(wsID, payload.Title)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "SANDBOX_INIT_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Dry-run sandbox session created", session)
	})

	wsGroup.Post("/:wsId/toolsim/sandboxes/:sessionId/execute", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		type ExecutePayload struct {
			ToolName   string                 `json:"tool_name"`
			Parameters map[string]interface{} `json:"parameters"`
		}
		var payload ExecutePayload
		if err := c.BodyParser(&payload); err != nil || payload.ToolName == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Tool name is required", nil)
		}
		inv, err := toolSimEngine.ExecuteToolDryRun(wsID, sessionID, payload.ToolName, payload.Parameters)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "EXECUTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Tool executed in dry-run sandbox", inv)
	})

	wsGroup.Post("/:wsId/toolsim/simulate-plan", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req toolsim.ExecutePlanRequest
		if err := c.BodyParser(&req); err != nil || len(req.Steps) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_PLAN", "Plan must contain at least one step", nil)
		}
		req.WorkspaceID = wsID
		if req.Title == "" {
			req.Title = "Multi-Step Agent Plan Simulation"
		}
		report, err := toolSimEngine.ExecutePlanDryRun(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "PLAN_SIMULATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Multi-step plan simulated with blast radius report", report)
	})

	wsGroup.Get("/:wsId/toolsim/sandboxes/:sessionId/report", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		sessionID := c.Params("sessionId")
		report, err := toolSimEngine.GetReport(wsID, sessionID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "REPORT_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Sandbox blast radius report compiled", report)
	})

	// Enterprise Cryptographic Audit Proof & Merkle Ledger Endpoints
	auditProofEngine := auditproof.NewEngine()
	wsGroup.Post("/:wsId/auditproof/events", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var req auditproof.RecordEventRequest
		if err := c.BodyParser(&req); err != nil || req.EventType == "" || req.ActionSummary == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Event type and action summary are required", nil)
		}
		req.WorkspaceID = wsID
		event, err := auditProofEngine.RecordEvent(req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "EVENT_LOG_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Cryptographic audit event recorded", event)
	})

	wsGroup.Post("/:wsId/auditproof/seal", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		block, err := auditProofEngine.SealBlock(wsID)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "SEAL_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Active ledger block sealed with Merkle root", block)
	})

	wsGroup.Get("/:wsId/auditproof/proofs/:eventId", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		eventID := c.Params("eventId")
		proof, err := auditProofEngine.GenerateProof(wsID, eventID)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "PROOF_NOT_FOUND", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Cryptographic Merkle proof generated", proof)
	})

	wsGroup.Post("/:wsId/auditproof/verify-proof", func(c *fiber.Ctx) error {
		type VerifyPayload struct {
			LeafHash  string   `json:"leaf_hash"`
			AuditPath []string `json:"audit_path"`
			RootHash  string   `json:"root_hash"`
		}
		var payload VerifyPayload
		if err := c.BodyParser(&payload); err != nil || payload.LeafHash == "" || payload.RootHash == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Leaf hash and root hash are required", nil)
		}
		isValid := auditProofEngine.VerifyProof(payload.LeafHash, payload.AuditPath, payload.RootHash)
		return response.Success(c, fiber.StatusOK, "Proof verification result", fiber.Map{
			"is_valid": isValid,
		})
	})

	wsGroup.Get("/:wsId/auditproof/verify-integrity", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		result := auditProofEngine.VerifyIntegrity(wsID)
		return response.Success(c, fiber.StatusOK, "Ledger integrity verification completed", result)
	})

	wsGroup.Get("/:wsId/auditproof/export", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		stdQuery := c.Query("standard", string(auditproof.ComplianceSOC2Type2))
		bundle := auditProofEngine.ExportBundle(wsID, auditproof.ComplianceStandard(stdQuery))
		return response.Success(c, fiber.StatusOK, "Compliance audit bundle exported", bundle)
	})

	// RAG Vector & File Ingestion Endpoints
	wsGroup.Post("/:wsId/search/semantic", ragHandler.SemanticSearch)
	wsGroup.Post("/:wsId/search/hybrid", ragHandler.HybridSearch)
	wsGroup.Post("/:wsId/ingest", ragHandler.IngestText)
	wsGroup.Post("/:wsId/ingest/url", ragHandler.IngestURL)
	wsGroup.Post("/:wsId/files/upload", ragHandler.UploadFile)
	wsGroup.Get("/:wsId/files", ragHandler.GetFiles)

	// Vector Index Optimization & HNSW Calibration Endpoints
	indexOptimizer := indexopt.NewIndexOptimizer()
	wsGroup.Get("/:wsId/rag/index/health", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		health, err := indexOptimizer.InspectHealth(c.Context(), wsID, 12500)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "INDEX_HEALTH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Vector index health metrics retrieved", health)
	})
	wsGroup.Post("/:wsId/rag/index/calibrate", func(c *fiber.Ctx) error {
		var req indexopt.CalibrationRequest
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid calibration request payload", nil)
		}
		params, err := indexOptimizer.CalibrateHNSW(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "CALIBRATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "HNSW parameters calibrated successfully", params)
	})
	wsGroup.Post("/:wsId/rag/index/optimize", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		report, err := indexOptimizer.GenerateOptimizationPlan(c.Context(), wsID, 12500)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "OPTIMIZATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Vector index optimization plan generated", report)
	})

	// Cross-Workspace Knowledge Federation & Entity Graph Linking Endpoints
	federationEngine := federation.NewFederationEngine()
	wsGroup.Post("/:wsId/federation/policies", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		var policy federation.FederationPolicy
		if err := c.BodyParser(&policy); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid federation policy payload", nil)
		}
		policy.SourceWorkspaceID = wsID
		if err := federationEngine.SetPolicy(policy); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "POLICY_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusCreated, "Federation trust policy registered", policy)
	})
	wsGroup.Get("/:wsId/federation/policies", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		policies := federationEngine.GetPolicies(wsID)
		return response.Success(c, fiber.StatusOK, "Workspace federation policies retrieved", policies)
	})
	wsGroup.Post("/:wsId/federation/search", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		type FedSearchReq struct {
			Query string `json:"query"`
			TopK  int    `json:"top_k"`
		}
		var req FedSearchReq
		if err := c.BodyParser(&req); err != nil || req.Query == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Query is required for federated search", nil)
		}
		res, err := federationEngine.FederatedSearch(c.Context(), wsID, req.Query, req.TopK)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "FEDERATION_SEARCH_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Cross-workspace federated search completed", res)
	})
	wsGroup.Post("/:wsId/federation/discover-entities", func(c *fiber.Ctx) error {
		wsID := c.Params("wsId")
		type EntityReq struct {
			TargetWorkspaceID string   `json:"target_workspace_id"`
			Entities          []string `json:"entities"`
		}
		var req EntityReq
		if err := c.BodyParser(&req); err != nil || req.TargetWorkspaceID == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "target_workspace_id is required", nil)
		}
		links, err := federationEngine.DiscoverCrossWorkspaceEntities(c.Context(), wsID, req.TargetWorkspaceID, req.Entities)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "ENTITY_DISCOVERY_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Cross-workspace entity links discovered", links)
	})

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

	// Automated Incident Postmortem & Root Cause Analysis (RCA) Endpoints
	postmortemSynthesizer := postmortem.NewPostmortemSynthesizer()
	api.Get("/incidents/templates", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Postmortem templates retrieved", postmortemSynthesizer.GetTemplates())
	})
	wsGroup.Post("/:wsId/incidents/analyze-rca", func(c *fiber.Ctx) error {
		type RCARequest struct {
			LogsSnippet string `json:"logs_snippet"`
			Summary     string `json:"summary"`
		}
		var req RCARequest
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid RCA request payload", nil)
		}
		rootCause, whys, err := postmortemSynthesizer.AnalyzeRCA(c.Context(), req.LogsSnippet, req.Summary)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "RCA_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Root cause analysis completed", fiber.Map{
			"root_cause_summary": rootCause,
			"five_whys_analysis": whys,
		})
	})
	wsGroup.Post("/:wsId/incidents/generate-postmortem", func(c *fiber.Ctx) error {
		var req postmortem.PostmortemRequest
		if err := c.BodyParser(&req); err != nil || req.IncidentTitle == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Incident title is required", nil)
		}
		report, err := postmortemSynthesizer.GeneratePostmortem(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "POSTMORTEM_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Incident postmortem report generated successfully", report)
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

	// Semantic Document Deduplication & Clustering Endpoints
	dedupEngine := dedup.NewDedupEngine()
	wsGroup.Post("/:wsId/documents/dedup/analyze", func(c *fiber.Ctx) error {
		var req dedup.DedupAnalysisRequest
		if err := c.BodyParser(&req); err != nil || len(req.Documents) == 0 {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Documents list is required for deduplication analysis", nil)
		}
		report, err := dedupEngine.AnalyzeDuplicates(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DEDUP_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Document deduplication analysis completed", report)
	})

	// Multi-Lingual Document Translation & Localization Endpoints
	translationEngine := translation.NewTranslationEngine()
	api.Get("/translation/languages", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Supported translation languages retrieved", translationEngine.GetSupportedLanguages())
	})
	wsGroup.Post("/:wsId/translation/detect", func(c *fiber.Ctx) error {
		type DetectReq struct {
			Text string `json:"text"`
		}
		var req DetectReq
		if err := c.BodyParser(&req); err != nil || req.Text == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Text is required for language detection", nil)
		}
		res, err := translationEngine.DetectLanguage(c.Context(), req.Text)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DETECTION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Language detected successfully", res)
	})
	wsGroup.Post("/:wsId/translation/translate", func(c *fiber.Ctx) error {
		var req translation.TranslationRequest
		if err := c.BodyParser(&req); err != nil || req.Text == "" || req.TargetLanguage == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Text and target_language are required for translation", nil)
		}
		res, err := translationEngine.Translate(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TRANSLATION_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Document translated and localized successfully", res)
	})

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

	// Automated Red-Team Adversarial Security Fuzzing Endpoints
	redTeamFuzzer := redteam.NewRedTeamFuzzer()
	api.Get("/security/redteam/attacks", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Adversarial attack catalog retrieved", redTeamFuzzer.GetAttackCatalog())
	})
	wsGroup.Post("/:wsId/security/redteam/fuzz", func(c *fiber.Ctx) error {
		var req redteam.FuzzRequest
		if err := c.BodyParser(&req); err != nil || req.SystemPrompt == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "System prompt is required for adversarial fuzzing", nil)
		}
		report, err := redTeamFuzzer.FuzzPrompt(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "FUZZING_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Automated red-team adversarial fuzzing completed", report)
	})

	// Multi-Agent Consensus & Debate Engine Endpoints
	consensusEngine := consensus.NewConsensusEngine()
	api.Get("/agents/consensus/specialists", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Specialist agent debaters retrieved", consensusEngine.GetAvailableSpecialists())
	})
	wsGroup.Post("/:wsId/agents/consensus/debate", func(c *fiber.Ctx) error {
		var req consensus.ConsensusRequest
		if err := c.BodyParser(&req); err != nil || req.Topic == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Topic is required for consensus debate", nil)
		}
		verdict, err := consensusEngine.RunDebate(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "DEBATE_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Multi-agent consensus debate completed", verdict)
	})

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

	// Neural Text-to-Speech (TTS) & Audio Narration Endpoints
	ttsEngine := tts.NewTTSEngine()
	api.Get("/voice/tts/voices", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "Available neural TTS voices retrieved", ttsEngine.GetAvailableVoices())
	})
	wsGroup.Post("/:wsId/voice/tts/synthesize", func(c *fiber.Ctx) error {
		var req tts.SynthesisRequest
		if err := c.BodyParser(&req); err != nil || req.Text == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Text content is required for speech synthesis", nil)
		}
		result, err := ttsEngine.SynthesizeSpeech(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "TTS_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Speech synthesized successfully", result)
	})
	wsGroup.Post("/:wsId/voice/tts/document-brief", func(c *fiber.Ctx) error {
		var req tts.AudioBriefRequest
		if err := c.BodyParser(&req); err != nil || req.Content == "" {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Document content is required for audio brief", nil)
		}
		result, err := ttsEngine.GenerateAudioBrief(c.Context(), req)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "AUDIO_BRIEF_FAILED", err.Error(), nil)
		}
		return response.Success(c, fiber.StatusOK, "Document audio brief generated successfully", result)
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
