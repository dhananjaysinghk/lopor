package memory

import (
	"errors"
	"math"
	"sync"
	"time"
)

// MemoryStore maintains long-term episodic memories with thread safety.
type MemoryStore struct {
	mu           sync.RWMutex
	items        map[string]MemoryItem
	userIndex    map[string][]string // workspaceID:userID -> list of memory IDs
	halfLifeDays float64
}

// NewMemoryStore initializes a memory storage layer.
func NewMemoryStore(halfLifeDays float64) *MemoryStore {
	if halfLifeDays <= 0 {
		halfLifeDays = 14.0 // Default 14-day half-life for episodic recency decay
	}
	return &MemoryStore{
		items:        make(map[string]MemoryItem),
		userIndex:    make(map[string][]string),
		halfLifeDays: halfLifeDays,
	}
}

func userKey(wsID, uID string) string {
	return wsID + ":" + uID
}

// Save stores or updates an episodic memory item.
func (s *MemoryStore) Save(item MemoryItem) MemoryItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if item.LastAccessedAt.IsZero() {
		item.LastAccessedAt = item.CreatedAt
	}
	if item.AccessCount <= 0 {
		item.AccessCount = 1
	}

	s.items[item.ID] = item

	k := userKey(item.WorkspaceID, item.UserID)
	ids := s.userIndex[k]
	exists := false
	for _, existingID := range ids {
		if existingID == item.ID {
			exists = true
			break
		}
	}
	if !exists {
		s.userIndex[k] = append(ids, item.ID)
	}

	return item
}

// Get retrieves a specific memory by its ID.
func (s *MemoryStore) Get(id string) (MemoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return MemoryItem{}, errors.New("memory item not found")
	}
	return item, nil
}

// IncrementAccess updates the access count and last accessed timestamp.
func (s *MemoryStore) IncrementAccess(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, ok := s.items[id]; ok {
		item.AccessCount++
		item.LastAccessedAt = time.Now().UTC()
		s.items[id] = item
	}
}

// List returns active memories for a given user in a workspace, optionally filtered by category.
func (s *MemoryStore) List(workspaceID, userID string, categories []MemoryCategory) []MemoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k := userKey(workspaceID, userID)
	ids := s.userIndex[k]

	categoryMap := make(map[MemoryCategory]bool)
	for _, c := range categories {
		categoryMap[c] = true
	}

	var results []MemoryItem
	for _, id := range ids {
		item, ok := s.items[id]
		if !ok || !item.IsActive {
			continue
		}
		if len(categories) > 0 && !categoryMap[item.Category] {
			continue
		}
		results = append(results, item)
	}

	return results
}

// Delete removes a specific memory item.
func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return false
	}

	delete(s.items, id)

	k := userKey(item.WorkspaceID, item.UserID)
	ids := s.userIndex[k]
	newIDs := make([]string, 0, len(ids))
	for _, existingID := range ids {
		if existingID != id {
			newIDs = append(newIDs, existingID)
		}
	}
	s.userIndex[k] = newIDs
	return true
}

// PurgeUser permanently removes all memories for a user (GDPR Right-to-be-Forgotten).
func (s *MemoryStore) PurgeUser(workspaceID, userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := userKey(workspaceID, userID)
	ids := s.userIndex[k]
	purged := 0

	for _, id := range ids {
		delete(s.items, id)
		purged++
	}

	delete(s.userIndex, k)
	return purged
}

// ApplyDecay recalculates decay scores across all memories based on time and access frequency.
func (s *MemoryStore) ApplyDecay() {
	s.mu.Lock()
	defer s.mu.Unlock()

	lambda := math.Ln2 / s.halfLifeDays
	now := time.Now().UTC()

	for id, item := range s.items {
		daysSinceAccess := now.Sub(item.LastAccessedAt).Hours() / 24.0
		if daysSinceAccess < 0 {
			daysSinceAccess = 0
		}

		effectiveLambda := lambda
		// Critical constraints decay at a fraction (5%) of standard memories to prevent forgetting
		if item.Category == CategoryConstraint && item.Importance >= 0.9 {
			effectiveLambda = lambda * 0.05
		}

		// Frequency boost dampens decay: (1 + 0.1 * ln(1 + AccessCount))
		frequencyBoost := 1.0 + 0.1*math.Log(1.0+float64(item.AccessCount))

		decayScore := item.Importance * math.Exp(-effectiveLambda*daysSinceAccess) * frequencyBoost
		if decayScore > 1.0 {
			decayScore = 1.0
		}
		if decayScore < 0.0 {
			decayScore = 0.0
		}

		item.DecayScore = math.Round(decayScore*1000) / 1000

		// Prune non-critical memories that decayed below threshold (0.10)
		if item.DecayScore < 0.10 && item.Category != CategoryConstraint {
			item.IsActive = false
		}

		s.items[id] = item
	}
}

// Count returns the number of active memories for a given user.
func (s *MemoryStore) Count(workspaceID, userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	k := userKey(workspaceID, userID)
	ids := s.userIndex[k]
	count := 0
	for _, id := range ids {
		if item, ok := s.items[id]; ok && item.IsActive {
			count++
		}
	}
	return count
}
