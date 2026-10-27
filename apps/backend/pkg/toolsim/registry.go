package toolsim

import (
	"strings"
	"sync"
)

// ToolRegistry manages the registration of system and workspace-custom mock tools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]MockTool // key: "workspace:tool_name" or "system:tool_name"
}

// NewToolRegistry initializes a registry populated with enterprise default mock tools.
func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		tools: make(map[string]MockTool),
	}
	r.registerBuiltInTools()
	return r
}

func (r *ToolRegistry) registerBuiltInTools() {
	builtIns := []MockTool{
		{
			Name:        "stripe_charge",
			Category:    CategoryPayment,
			RiskLevel:   RiskHigh,
			Description: "Simulates an enterprise credit card transaction through Stripe API.",
			ParametersSchema: map[string]string{
				"amount":      "float (amount in currency units)",
				"currency":    "string (e.g. usd, eur)",
				"customer_id": "string (e.g. cus_99182)",
			},
			DefaultMockResponse: map[string]interface{}{
				"status":    "succeeded",
				"charge_id": "ch_mock_stripe_9921",
				"paid":      true,
			},
			SimulatedLatencyMs: 140,
		},
		{
			Name:        "aws_s3_delete_bucket",
			Category:    CategoryInfrastructure,
			RiskLevel:   RiskCritical,
			Description: "Simulates permanent deletion of an AWS S3 cloud storage bucket.",
			ParametersSchema: map[string]string{
				"bucket_name": "string (bucket identifier)",
				"force":       "boolean (force delete all objects)",
			},
			DefaultMockResponse: map[string]interface{}{
				"deleted":        true,
				"objects_purged": 142,
			},
			SimulatedLatencyMs: 220,
		},
		{
			Name:        "postgres_exec_query",
			Category:    CategoryDatabase,
			RiskLevel:   RiskMedium,
			Description: "Simulates execution of SQL queries against relational databases.",
			ParametersSchema: map[string]string{
				"query":       "string (SQL query text)",
				"timeout_sec": "integer (query timeout)",
			},
			DefaultMockResponse: map[string]interface{}{
				"status":        "COMMAND_OK",
				"rows_affected": 3,
			},
			SimulatedLatencyMs: 45,
		},
		{
			Name:        "slack_post_message",
			Category:    CategoryCommunication,
			RiskLevel:   RiskLow,
			Description: "Simulates dispatching a notification message to a Slack channel.",
			ParametersSchema: map[string]string{
				"channel": "string (channel name or ID)",
				"text":    "string (message content)",
			},
			DefaultMockResponse: map[string]interface{}{
				"ok":      true,
				"ts":      "1710000000.001",
				"channel": "general",
			},
			SimulatedLatencyMs: 65,
		},
		{
			Name:        "github_merge_pr",
			Category:    CategoryInfrastructure,
			RiskLevel:   RiskMedium,
			Description: "Simulates merging a GitHub Pull Request into the primary branch.",
			ParametersSchema: map[string]string{
				"repo":         "string (owner/repo)",
				"pull_number":  "integer (PR number)",
				"commit_title": "string (commit message)",
			},
			DefaultMockResponse: map[string]interface{}{
				"merged":  true,
				"sha":     "93ee3ae714mock",
				"message": "Pull Request successfully merged",
			},
			SimulatedLatencyMs: 110,
		},
	}

	for _, tool := range builtIns {
		key := "system:" + strings.ToLower(tool.Name)
		r.tools[key] = tool
	}
}

// Register registers or overrides a mock tool for a specific workspace.
func (r *ToolRegistry) Register(tool MockTool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := "system"
	if tool.WorkspaceID != "" {
		prefix = tool.WorkspaceID
	}
	key := prefix + ":" + strings.ToLower(tool.Name)
	r.tools[key] = tool
}

// Get finds a tool for the workspace, falling back to system defaults.
func (r *ToolRegistry) Get(workspaceID, toolName string) (*MockTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowerName := strings.ToLower(toolName)

	// Workspace specific check
	if workspaceID != "" {
		wsKey := workspaceID + ":" + lowerName
		if tool, exists := r.tools[wsKey]; exists {
			return &tool, true
		}
	}

	// System default fallback
	sysKey := "system:" + lowerName
	if tool, exists := r.tools[sysKey]; exists {
		return &tool, true
	}

	return nil, false
}

// List returns all tools available to the given workspace.
func (r *ToolRegistry) List(workspaceID string) []MockTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var result []MockTool

	// 1. Workspace custom tools
	for k, tool := range r.tools {
		if strings.HasPrefix(k, workspaceID+":") {
			result = append(result, tool)
			seen[tool.Name] = true
		}
	}

	// 2. System tools (if not shadowed)
	for k, tool := range r.tools {
		if strings.HasPrefix(k, "system:") && !seen[tool.Name] {
			result = append(result, tool)
		}
	}

	return result
}
