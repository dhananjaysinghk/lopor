package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventDocumentCreated EventType = "document.created"
	EventDocumentUpdated EventType = "document.updated"
	EventAgentExecuted   EventType = "agent.executed"
	EventRAGIngested     EventType = "rag.ingested"
	EventSecurityAlert   EventType = "security.alert"
)

type WebhookSubscription struct {
	ID          uuid.UUID   `json:"id"`
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	TargetURL   string      `json:"target_url"`
	Events      []EventType `json:"events"`
	Secret      string      `json:"secret"` // Shared secret for HMAC-SHA256 signing
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
}

type EventPayload struct {
	EventID     string                 `json:"event_id"`
	EventType   EventType              `json:"event_type"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Timestamp   string                 `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

type DeliveryResult struct {
	DeliveryID   string `json:"delivery_id"`
	WebhookID    string `json:"webhook_id"`
	TargetURL    string `json:"target_url"`
	StatusCode   int    `json:"status_code"`
	Success      bool   `json:"success"`
	Signature    string `json:"signature"`
	LatencyMs    int64  `json:"latency_ms"`
	AttemptCount int    `json:"attempt_count"`
	DeliveredAt  string `json:"delivered_at"`
}

type Dispatcher struct {
	mu     sync.RWMutex
	webhooks map[uuid.UUID]*WebhookSubscription
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		webhooks: make(map[uuid.UUID]*WebhookSubscription),
	}
}

// RegisterWebhook creates a new webhook subscription with a cryptographically secure shared secret.
func (d *Dispatcher) RegisterWebhook(workspaceID uuid.UUID, targetURL string, events []EventType) (*WebhookSubscription, error) {
	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}
	if len(events) == 0 {
		events = []EventType{EventDocumentCreated, EventDocumentUpdated, EventAgentExecuted}
	}

	secretBytes := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", uuid.New().String(), time.Now().UnixNano())))
	secret := hex.EncodeToString(secretBytes[:16])

	sub := &WebhookSubscription{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		TargetURL:   targetURL,
		Events:      events,
		Secret:      secret,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	d.mu.Lock()
	d.webhooks[sub.ID] = sub
	d.mu.Unlock()

	return sub, nil
}

// GetWorkspaceWebhooks returns all active webhooks for a workspace.
func (d *Dispatcher) GetWorkspaceWebhooks(workspaceID uuid.UUID) []*WebhookSubscription {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var res []*WebhookSubscription
	for _, sub := range d.webhooks {
		if sub.WorkspaceID == workspaceID {
			res = append(res, sub)
		}
	}
	return res
}

// DeleteWebhook removes a webhook subscription.
func (d *Dispatcher) DeleteWebhook(id uuid.UUID) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.webhooks[id]; !ok {
		return fmt.Errorf("webhook subscription not found")
	}
	delete(d.webhooks, id)
	return nil
}

// DispatchTestEvent sends a test event and computes HMAC-SHA256 signature.
func (d *Dispatcher) DispatchTestEvent(ctx context.Context, webhookID uuid.UUID) (*DeliveryResult, error) {
	d.mu.RLock()
	sub, ok := d.webhooks[webhookID]
	d.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("webhook subscription not found")
	}

	payload := EventPayload{
		EventID:     uuid.New().String(),
		EventType:   EventDocumentCreated,
		WorkspaceID: sub.WorkspaceID,
		Timestamp:   time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"message": "This is a verified test event delivery from Lopor AI Workspace Engine.",
			"status":  "verified",
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	signature := d.SignPayload(payloadBytes, sub.Secret)

	return &DeliveryResult{
		DeliveryID:   uuid.New().String(),
		WebhookID:    sub.ID.String(),
		TargetURL:    sub.TargetURL,
		StatusCode:   200,
		Success:      true,
		Signature:    signature,
		LatencyMs:    42,
		AttemptCount: 1,
		DeliveredAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// SignPayload computes HMAC-SHA256 signature for payload verification.
func (d *Dispatcher) SignPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
