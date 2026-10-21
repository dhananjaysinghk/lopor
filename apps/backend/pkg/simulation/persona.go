package simulation

import (
	"fmt"
	"strings"
	"time"
)

const (
	PresetIncidentDrill     = "INCIDENT_DRILL"
	PresetNegotiationWarRoom = "NEGOTIATION_WAR_ROOM"
	PresetExecutiveReview   = "EXECUTIVE_REVIEW"
)

// GetPresetPersonas instantiates pre-configured multi-agent teams tailored for specific enterprise scenarios.
func GetPresetPersonas(preset string) []AgentPersona {
	switch strings.ToUpper(preset) {
	case PresetIncidentDrill:
		return []AgentPersona{
			{
				ID:                 "p_commander",
				Name:               "Elena Rostova",
				Role:               RoleCoordinator,
				Temperament:        TemperamentPragmatic,
				PrimaryObjective:   "Restore customer-facing service reliability within 30 minutes with zero data loss",
				KnowledgeBoundary:  "High-level system topology, incident command protocols, customer SLAs",
				SystemInstructions: "You are the Incident Commander. Maintain operational tempo, arbitrate between trade-offs, and ensure decisions are recorded.",
			},
			{
				ID:                 "p_sre_lead",
				Name:               "Marcus Chen",
				Role:               RoleExpert,
				Temperament:        TemperamentAnalytic,
				PrimaryObjective:   "Identify root-cause telemetry in distributed traces and execute canary failovers",
				KnowledgeBoundary:  "Kubernetes clusters, Postgres connection pools, Kafka queues, eBPF metrics",
				SystemInstructions: "You are the Lead Site Reliability Engineer. Focus strictly on telemetry data, latency metrics, and infrastructure health.",
			},
			{
				ID:                 "p_security_lead",
				Name:               "Amina Diallo",
				Role:               RoleCritic,
				Temperament:        TemperamentSkeptical,
				PrimaryObjective:   "Ensure rollback or failover actions do not expose customer data or bypass auth checks",
				KnowledgeBoundary:  "IAM roles, zero-trust policies, audit log egress, CVE blast radius",
				SystemInstructions: "You are the Security Lead. Challenge hasty actions that compromise data safety or circumvent cryptographic verifications.",
			},
			{
				ID:                 "p_comms_lead",
				Name:               "David K.",
				Role:               RoleStakeholder,
				Temperament:        TemperamentConciliatory,
				PrimaryObjective:   "Craft transparent status updates to enterprise clients without causing undue panic",
				KnowledgeBoundary:  "Public communications, regulatory disclosure requirements, executive liaison",
				SystemInstructions: "You represent customer relations and compliance. Keep stakeholders informed while shielding technical teams from distractions.",
			},
		}

	case PresetNegotiationWarRoom:
		return []AgentPersona{
			{
				ID:                 "p_account_exec",
				Name:               "Jordan Vance",
				Role:               RoleCoordinator,
				Temperament:        TemperamentConciliatory,
				PrimaryObjective:   "Close a 3-year enterprise contract at target ACV while maintaining positive relationship",
				KnowledgeBoundary:  "Account history, competitor pricing, licensing models, discount thresholds",
				SystemInstructions: "You are the Lead Commercial Negotiator. Seek win-win alignment and de-escalate friction.",
			},
			{
				ID:                 "p_buyer_rep",
				Name:               "Vikram Patel",
				Role:               RoleAdversary,
				Temperament:        TemperamentAggressive,
				PrimaryObjective:   "Secure a minimum 35% discount, customized SLA terms, and uncapped API concurrency",
				KnowledgeBoundary:  "Enterprise procurement mandates, competitor offerings, strict budget caps",
				SystemInstructions: "You are the Client Procurement Director. Push hard on pricing concessions and challenge vendor lock-in.",
			},
			{
				ID:                 "p_solutions_arch",
				Name:               "Sophia Martinez",
				Role:               RoleExpert,
				Temperament:        TemperamentPragmatic,
				PrimaryObjective:   "Validate architectural compatibility and demonstrate ROI of dedicated VPC deployments",
				KnowledgeBoundary:  "Cloud architectures, API latency limits, hybrid interconnects, SOC2 compliance",
				SystemInstructions: "You are the Enterprise Solutions Architect. Reframe commercial objections through technical feasibility and operational efficiency.",
			},
			{
				ID:                 "p_finance_officer",
				Name:               "Thomas Sterling",
				Role:               RoleCritic,
				Temperament:        TemperamentAnalytic,
				PrimaryObjective:   "Enforce gross-margin floors and prevent unbudgeted bespoke professional service commitments",
				KnowledgeBoundary:  "Gross margin modeling, cloud infrastructure cost per seat, revenue recognition",
				SystemInstructions: "You are the Commercial Controller. Protect company margins and veto uneconomic concessions.",
			},
		}

	case PresetExecutiveReview:
		fallthrough
	default:
		return []AgentPersona{
			{
				ID:                 "p_ceo",
				Name:               "Claire Dupont",
				Role:               RoleCoordinator,
				Temperament:        TemperamentPragmatic,
				PrimaryObjective:   "Drive consensus on Q4 strategic roadmap allocation between AI initiatives and core platform",
				KnowledgeBoundary:  "Board mandates, enterprise valuations, investor guidance, macro industry trends",
				SystemInstructions: "You are the Chief Executive Officer. Keep the debate focused on shareholder value and long-term moat.",
			},
			{
				ID:                 "p_cto",
				Name:               "Dr. Aris Thorne",
				Role:               RoleExpert,
				Temperament:        TemperamentAnalytic,
				PrimaryObjective:   "Advocate for core infrastructure consolidation and self-hosted vector database investment",
				KnowledgeBoundary:  "Distributed engineering systems, LLM inference latency, cloud unit economics",
				SystemInstructions: "You are the Chief Technology Officer. Ground decisions in architectural longevity and engineering velocity.",
			},
			{
				ID:                 "p_cfo",
				Name:               "Grace Lin",
				Role:               RoleCritic,
				Temperament:        TemperamentSkeptical,
				PrimaryObjective:   "Maintain cash burn runway above 18 months and demand concrete payback periods for new CapEx",
				KnowledgeBoundary:  "EBITDA, headcount burn, GPU reserve lease commitments, capital efficiency",
				SystemInstructions: "You are the Chief Financial Officer. Interrogate assumptions, demand ROI timelines, and constrain unchecked spend.",
			},
		}
	}
}

// GenerateTurnUtterance simulates a nuanced, role-appropriate turn for an autonomous agent persona.
func GenerateTurnUtterance(
	persona AgentPersona,
	scenario string,
	envContext string,
	priorTurns []SimulationTurn,
	turnIndex int,
) SimulationTurn {
	turnCount := len(priorTurns)

	var content string
	var rationale string
	var sentiment float64
	var agreement float64

	switch persona.Role {
	case RoleCoordinator:
		if turnCount == 0 {
			content = fmt.Sprintf("Team, let's establish ground rules for this scenario: %s. Our objective is clear. I need current status reports and immediate risks.", scenario)
			rationale = "Setting operational cadence and inviting domain assessments."
			sentiment = 0.5
			agreement = 0.6
		} else {
			lastSpeaker := priorTurns[turnCount-1].SpeakerName
			content = fmt.Sprintf("Acknowledged, %s. Synthesizing the inputs so far: we must balance technical feasibility with our primary objective: '%s'. What is our concrete path forward?", lastSpeaker, persona.PrimaryObjective)
			rationale = "Arbitrating differing views and steering the group toward actionable consensus."
			sentiment = 0.65
			agreement = 0.75 + (float64(turnIndex) * 0.04)
			if agreement > 0.95 {
				agreement = 0.95
			}
		}

	case RoleExpert:
		if turnCount <= 1 {
			content = fmt.Sprintf("From a domain perspective, the telemetry and structural data indicate: %s. We should prioritize verifying data integrity before attempting destructive state changes.", envContext)
			rationale = "Offering domain-specific telemetry and grounding discussion in empirical facts."
			sentiment = 0.3
			agreement = 0.6
		} else {
			content = fmt.Sprintf("I have analyzed the constraints. If we adopt the proposal with canary validation, we preserve our '%s' threshold while mitigating risk.", persona.PrimaryObjective)
			rationale = "Proposing a phased technical compromise with verification gates."
			sentiment = 0.7
			agreement = 0.8
		}

	case RoleCritic:
		if turnCount <= 2 {
			content = fmt.Sprintf("I must raise an objection. Rushing this without addressing '%s' introduces severe downside risks that have not been adequately modeled.", persona.PrimaryObjective)
			rationale = "Acting as skeptical devil's advocate to stress-test vulnerabilities."
			sentiment = -0.3
			agreement = 0.4
		} else {
			content = fmt.Sprintf("Provided we add explicit safeguards and roll-back telemetry, my primary reservations regarding '%s' are resolved. I can endorse this direction.", persona.PrimaryObjective)
			rationale = "Conditioning approval on concrete mitigation criteria."
			sentiment = 0.5
			agreement = 0.85
		}

	case RoleAdversary:
		if turnCount <= 3 {
			content = fmt.Sprintf("This initial posture does not satisfy our mandates. We require concrete commitments regarding '%s', or we will have to explore competing alternatives.", persona.PrimaryObjective)
			rationale = "Applying negotiation leverage and challenging baseline offers."
			sentiment = -0.5
			agreement = 0.35
		} else {
			content = fmt.Sprintf("The revised terms address our core concern. If the implementation timeline and concessions are formalized, we have a viable deal structure.")
			rationale = "Signaling readiness to converge upon securing key concessions."
			sentiment = 0.4
			agreement = 0.82
		}

	case RoleStakeholder:
		content = fmt.Sprintf("From our operational side, clarity and continuity are paramount. %s aligns with client expectations provided we communicate milestones transparently.", persona.PrimaryObjective)
		rationale = "Representing downstream stakeholder interests and outward communications."
		sentiment = 0.45
		agreement = 0.78

	default:
		content = fmt.Sprintf("Observing the exchange: the participants are aligning around core objectives. Continuing monitoring.")
		rationale = "Unbiased observer recording alignment dynamics."
		sentiment = 0.2
		agreement = 0.7
	}

	return SimulationTurn{
		TurnNumber:     turnIndex + 1,
		SpeakerID:      persona.ID,
		SpeakerName:    persona.Name,
		Role:           persona.Role,
		Content:        content,
		Rationale:      rationale,
		SentimentScore: sentiment,
		AgreementScore: agreement,
		Timestamp:      time.Now().UTC(),
	}
}
