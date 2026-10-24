package temporalkg

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ExecuteTemporalQuery filters facts based on time-travel point snapshots, temporal intervals, or current state.
func ExecuteTemporalQuery(store *TemporalStore, req TemporalQueryRequest) TemporalQueryResponse {
	allFacts := store.GetAllFactsForWorkspace(req.WorkspaceID)
	var matched []TemporalFact

	for _, f := range allFacts {
		if !matchesEntityAndPredicate(f, req) {
			continue
		}

		if req.OnlyCurrent && !f.IsCurrent {
			continue
		}

		// Point-in-time "AsOf" snapshot query
		if req.AsOf != nil {
			asOf := *req.AsOf
			if f.ValidFrom.After(asOf) {
				continue
			}
			if f.ValidTo != nil && !f.ValidTo.After(asOf) {
				continue
			}
		}

		// Interval query [FromTime, ToTime]
		if req.FromTime != nil && req.ToTime != nil {
			from := *req.FromTime
			to := *req.ToTime
			// Fact ends before window
			if f.ValidTo != nil && f.ValidTo.Before(from) {
				continue
			}
			// Fact starts after window
			if f.ValidFrom.After(to) {
				continue
			}
		}

		matched = append(matched, *f)
	}

	// Sort by ValidFrom descending
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].ValidFrom.After(matched[j].ValidFrom)
	})

	groundedContext := SynthesizeTemporalContext(matched, req.AsOf)

	return TemporalQueryResponse{
		WorkspaceID:     req.WorkspaceID,
		MatchedFacts:    matched,
		AsOfTimestamp:   req.AsOf,
		GroundedContext: groundedContext,
		TotalMatches:    len(matched),
	}
}

// BuildEntityEvolution reconstructs the full chronological history of state mutations for an entity.
func BuildEntityEvolution(store *TemporalStore, workspaceID, entityID string) (*EntityEvolution, error) {
	entity, err := store.GetEntity(workspaceID, entityID)
	if err != nil {
		return nil, err
	}

	allFacts := store.GetAllFactsForWorkspace(workspaceID)
	var entityFacts []*TemporalFact
	for _, f := range allFacts {
		if f.SubjectID == entityID || f.ObjectID == entityID {
			entityFacts = append(entityFacts, f)
		}
	}

	var events []FactEvent
	var currentFacts []TemporalFact

	for _, f := range entityFacts {
		validInterval := fmt.Sprintf("[%s -> Ongoing]", f.ValidFrom.Format("2006-01-02"))
		if f.ValidTo != nil {
			validInterval = fmt.Sprintf("[%s -> %s]", f.ValidFrom.Format("2006-01-02"), f.ValidTo.Format("2006-01-02"))
		}

		// Creation event
		targetName := f.ObjectName
		if f.ObjectID == entityID {
			targetName = f.SubjectName
		}

		events = append(events, FactEvent{
			Timestamp:      f.ValidFrom,
			Mutation:       MutationCreated,
			Predicate:      f.Predicate,
			TargetEntity:   targetName,
			ValidInterval:  validInterval,
			SourceCitation: f.SourceCitation,
		})

		// Superseded event if applicable
		if f.SupersededAt != nil {
			events = append(events, FactEvent{
				Timestamp:      *f.SupersededAt,
				Mutation:       MutationSuperseded,
				Predicate:      f.Predicate,
				TargetEntity:   targetName,
				ValidInterval:  validInterval,
				SourceCitation: f.SourceCitation,
			})
		}

		if f.IsCurrent {
			currentFacts = append(currentFacts, *f)
		}
	}

	// Sort events chronologically
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return &EntityEvolution{
		EntityID:      entity.ID,
		CanonicalName: entity.CanonicalName,
		Type:          entity.Type,
		Events:        events,
		CurrentFacts:  currentFacts,
	}, nil
}

// SynthesizeTemporalContext generates LLM-ready markdown groundings indicating temporal validity.
func SynthesizeTemporalContext(facts []TemporalFact, asOf *time.Time) string {
	if len(facts) == 0 {
		return "No temporal knowledge graph facts identified for the specified query and time window."
	}

	var b strings.Builder
	if asOf != nil {
		b.WriteString(fmt.Sprintf("### TEMPORAL KNOWLEDGE GRAPH EVIDENCE (POINT-IN-TIME AS OF %s)\n", asOf.Format("2006-01-02 15:04:05 UTC")))
	} else {
		b.WriteString("### TEMPORAL KNOWLEDGE GRAPH EVIDENCE (CURRENT STATE & HISTORICAL RECORD)\n")
	}

	for i, f := range facts {
		validity := "ACTIVE / CURRENT"
		if !f.IsCurrent || f.ValidTo != nil {
			if f.ValidTo != nil {
				validity = fmt.Sprintf("HISTORICAL (%s to %s)", f.ValidFrom.Format("2006-01-02"), f.ValidTo.Format("2006-01-02"))
			} else {
				validity = fmt.Sprintf("SUPERSEDED (Started %s)", f.ValidFrom.Format("2006-01-02"))
			}
		}

		b.WriteString(fmt.Sprintf("%d. **%s** `[%s]` **%s**\n", i+1, f.SubjectName, f.Predicate, f.ObjectName))
		b.WriteString(fmt.Sprintf("   - Validity Status: %s\n", validity))
		b.WriteString(fmt.Sprintf("   - Confidence: %.0f%%\n", f.Confidence*100))
		if f.SourceCitation != "" {
			b.WriteString(fmt.Sprintf("   - Provenance / Citation: %s\n", f.SourceCitation))
		}
	}

	return strings.TrimSpace(b.String())
}

func matchesEntityAndPredicate(f *TemporalFact, req TemporalQueryRequest) bool {
	if req.EntityID != "" && f.SubjectID != req.EntityID && f.ObjectID != req.EntityID {
		return false
	}
	if req.Predicate != "" && !strings.EqualFold(f.Predicate, req.Predicate) {
		return false
	}
	return true
}
