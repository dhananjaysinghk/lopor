package taskplanner_test

import (
	"context"
	"testing"

	"github.com/lopor-ai/lopor/pkg/taskplanner"
)

func TestTaskPlanner_ValidDAG(t *testing.T) {
	planner := taskplanner.NewTaskPlanner()

	nodes := []taskplanner.TaskNode{
		{ID: "task-1", Name: "Fetch Data", ActionType: "web_scrape"},
		{ID: "task-2", Name: "Embed Vectors", ActionType: "rag_search", Dependencies: []string{"task-1"}},
		{ID: "task-3", Name: "Summarize Findings", ActionType: "llm_generate", Dependencies: []string{"task-2"}},
		{ID: "task-4", Name: "Code Export", ActionType: "code_exec", Dependencies: []string{"task-1"}},
	}

	plan, err := planner.PlanDAG("Scrape and Summarize Pipeline", nodes)
	if err != nil {
		t.Fatalf("unexpected error planning DAG: %v", err)
	}

	if len(plan.Stages) == 0 {
		t.Fatalf("expected stages in plan, got 0")
	}

	// Stage 0 should be task-1
	if len(plan.Stages[0]) != 1 || plan.Stages[0][0] != "task-1" {
		t.Errorf("expected task-1 in initial stage, got %v", plan.Stages[0])
	}

	// Execute the plan
	res, err := planner.ExecuteDAG(context.Background(), plan)
	if err != nil {
		t.Fatalf("unexpected error executing DAG: %v", err)
	}

	if !res.Success {
		t.Errorf("expected execution success, got false")
	}

	if res.TotalTasks != 4 {
		t.Errorf("expected 4 completed tasks, got %d", res.TotalTasks)
	}
}

func TestTaskPlanner_CircularDependency(t *testing.T) {
	planner := taskplanner.NewTaskPlanner()

	// Circular cycle: task-A -> task-B -> task-A
	nodes := []taskplanner.TaskNode{
		{ID: "task-A", Name: "Task A", Dependencies: []string{"task-B"}},
		{ID: "task-B", Name: "Task B", Dependencies: []string{"task-A"}},
	}

	_, err := planner.PlanDAG("Circular Workflow", nodes)
	if err == nil {
		t.Errorf("expected error detecting circular dependency in DAG, got nil")
	}
}
