package webhook_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/pkg/webhook"
)

func TestWebhookDispatcher_Lifecycle(t *testing.T) {
	dispatcher := webhook.NewDispatcher()
	wsID := uuid.New()

	// 1. Register Webhook
	sub, err := dispatcher.RegisterWebhook(wsID, "https://api.example.com/webhook", []webhook.EventType{
		webhook.EventDocumentCreated,
		webhook.EventSecurityAlert,
	})
	if err != nil {
		t.Fatalf("unexpected error registering webhook: %v", err)
	}

	if sub.ID == uuid.Nil {
		t.Errorf("expected valid webhook UUID, got Nil")
	}

	if len(sub.Secret) == 0 {
		t.Errorf("expected non-empty shared secret for signing")
	}

	// 2. Fetch Workspace Webhooks
	list := dispatcher.GetWorkspaceWebhooks(wsID)
	if len(list) != 1 {
		t.Errorf("expected 1 registered webhook, got %d", len(list))
	}

	// 3. Dispatch Test Event
	res, err := dispatcher.DispatchTestEvent(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("failed to dispatch test event: %v", err)
	}

	if !res.Success {
		t.Errorf("expected test delivery success, got false")
	}

	if !strings.HasPrefix(res.Signature, "sha256=") {
		t.Errorf("expected HMAC-SHA256 signature prefix sha256=, got %s", res.Signature)
	}

	// 4. Delete Webhook
	err = dispatcher.DeleteWebhook(sub.ID)
	if err != nil {
		t.Fatalf("failed to delete webhook: %v", err)
	}

	listAfter := dispatcher.GetWorkspaceWebhooks(wsID)
	if len(listAfter) != 0 {
		t.Errorf("expected 0 webhooks after deletion, got %d", len(listAfter))
	}
}
