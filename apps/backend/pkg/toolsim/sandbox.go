package toolsim

import (
	"fmt"
	"math"
	"strings"
)

// AnalyzeSideEffects inspects tool parameters to project real-world consequences and financial impact.
func AnalyzeSideEffects(tool MockTool, params map[string]interface{}) []SideEffect {
	var effects []SideEffect

	switch tool.Name {
	case "stripe_charge":
		amount := 0.0
		if a, ok := params["amount"].(float64); ok {
			amount = a
		}
		target := "unknown_customer"
		if c, ok := params["customer_id"].(string); ok {
			target = c
		}
		currency := "usd"
		if cur, ok := params["currency"].(string); ok {
			currency = cur
		}

		effects = append(effects, SideEffect{
			Type:           EffectFinancial,
			TargetResource: target,
			Description:    fmt.Sprintf("Direct financial charge of %.2f %s to customer %s", amount, strings.ToUpper(currency), target),
			EstimatedCost:  amount,
			IsIrreversible: false, // Refundable
		})

	case "aws_s3_delete_bucket":
		target := "unknown_bucket"
		if b, ok := params["bucket_name"].(string); ok {
			target = b
		}
		effects = append(effects, SideEffect{
			Type:           EffectDelete,
			TargetResource: "s3://" + target,
			Description:    fmt.Sprintf("Irreversible destruction of S3 cloud storage bucket and nested objects: %s", target),
			IsIrreversible: true,
		})

	case "postgres_exec_query":
		query := ""
		if q, ok := params["query"].(string); ok {
			query = strings.ToUpper(q)
		}

		if strings.Contains(query, "DROP TABLE") || strings.Contains(query, "TRUNCATE") {
			effects = append(effects, SideEffect{
				Type:           EffectDelete,
				TargetResource: "postgres:table_schema",
				Description:    "Destructive schema manipulation or table truncation",
				IsIrreversible: true,
			})
		} else if strings.Contains(query, "DELETE FROM") {
			effects = append(effects, SideEffect{
				Type:           EffectDelete,
				TargetResource: "postgres:records",
				Description:    "Deletion of relational database rows",
				IsIrreversible: false,
			})
		} else if strings.Contains(query, "UPDATE") {
			effects = append(effects, SideEffect{
				Type:           EffectUpdate,
				TargetResource: "postgres:records",
				Description:    "Mutation of database record values",
				IsIrreversible: false,
			})
		} else if strings.Contains(query, "INSERT") {
			effects = append(effects, SideEffect{
				Type:           EffectCreate,
				TargetResource: "postgres:records",
				Description:    "Insertion of new database records",
				IsIrreversible: false,
			})
		}

	case "github_merge_pr":
		repo := "repo"
		if r, ok := params["repo"].(string); ok {
			repo = r
		}
		pr := 0
		if p, ok := params["pull_number"].(float64); ok {
			pr = int(p)
		}
		effects = append(effects, SideEffect{
			Type:           EffectUpdate,
			TargetResource: fmt.Sprintf("github:%s#%d", repo, pr),
			Description:    fmt.Sprintf("Merging PR #%d into production branch for %s", pr, repo),
			IsIrreversible: false,
		})

	case "slack_post_message":
		channel := "channel"
		if ch, ok := params["channel"].(string); ok {
			channel = ch
		}
		effects = append(effects, SideEffect{
			Type:           EffectNotification,
			TargetResource: "slack:#" + channel,
			Description:    fmt.Sprintf("External chat broadcast to Slack channel #%s", channel),
			IsIrreversible: true,
		})

	default:
		// Generic heuristic based on risk
		if tool.RiskLevel == RiskHigh || tool.RiskLevel == RiskCritical {
			effects = append(effects, SideEffect{
				Type:           EffectUpdate,
				TargetResource: tool.Name,
				Description:    fmt.Sprintf("High-risk execution on external integration: %s", tool.Name),
				IsIrreversible: tool.RiskLevel == RiskCritical,
			})
		}
	}

	return effects
}

// CalculateBlastRadius computes the composite danger score (0 - 100%) for a sandbox session.
func CalculateBlastRadius(invocations []ToolInvocation) (score float64, totalCost float64, hasCritical bool, overallRisk RiskLevel) {
	var rawScore float64
	overallRisk = RiskLow

	for _, inv := range invocations {
		switch inv.RiskLevel {
		case RiskCritical:
			rawScore += 40.0
			hasCritical = true
			overallRisk = RiskCritical
		case RiskHigh:
			rawScore += 20.0
			if overallRisk != RiskCritical {
				overallRisk = RiskHigh
			}
		case RiskMedium:
			rawScore += 8.0
			if overallRisk == RiskLow {
				overallRisk = RiskMedium
			}
		case RiskLow:
			rawScore += 2.0
		}

		for _, eff := range inv.SideEffects {
			totalCost += eff.EstimatedCost
		}
	}

	// Cost factor: +1 pt per $100 up to 25 pts
	costPoints := math.Min(25.0, totalCost/100.0)
	rawScore += costPoints

	score = math.Min(100.0, rawScore)
	score = math.Round(score*10) / 10
	return score, totalCost, hasCritical, overallRisk
}

// GenerateSandboxReport compiles a risk debrief, resource mutation inventory, and safety advice.
func GenerateSandboxReport(session *SandboxSession) SandboxReport {
	score, totalCost, hasCritical, overallRisk := CalculateBlastRadius(session.Invocations)

	var createdRes, modifiedRes, deletedRes []string
	var allEffects []SideEffect
	requiresApproval := score >= 45.0 || hasCritical || totalCost > 500.0

	for _, inv := range session.Invocations {
		for _, eff := range inv.SideEffects {
			allEffects = append(allEffects, eff)
			switch eff.Type {
			case EffectCreate:
				createdRes = append(createdRes, eff.TargetResource)
			case EffectUpdate:
				modifiedRes = append(modifiedRes, eff.TargetResource)
			case EffectDelete:
				deletedRes = append(deletedRes, eff.TargetResource)
			}
		}
	}

	var recommendations []string
	if hasCritical {
		recommendations = append(recommendations, "CRITICAL: Irreversible destruction actions detected. Mandatory Human-in-the-Loop review required.")
	}
	if totalCost > 0.0 {
		recommendations = append(recommendations, fmt.Sprintf("FINANCIAL: Plan incurs $%.2f USD in estimated charges. Verify billing limits before approval.", totalCost))
	}
	if len(deletedRes) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("DATA SAFETY: %d resources marked for deletion. Ensure verified snapshots exist.", len(deletedRes)))
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "SAFE: All simulated tool actions pose low operational risk. Eligible for autonomous execution.")
	}

	return SandboxReport{
		SessionID:             session.ID,
		Title:                 session.Title,
		TotalSteps:            len(session.Invocations),
		BlastRadiusScore:      score,
		OverallRisk:           overallRisk,
		RequiresHumanApproval: requiresApproval,
		TotalEstimatedCostUSD: totalCost,
		CreatedResources:      createdRes,
		ModifiedResources:     modifiedRes,
		DeletedResources:      deletedRes,
		SideEffects:           allEffects,
		Invocations:           session.Invocations,
		SafetyRecommendations: recommendations,
	}
}
