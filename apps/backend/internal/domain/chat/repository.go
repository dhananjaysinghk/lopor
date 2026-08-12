package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type Repository interface {
	CreateChat(ctx context.Context, chat *models.Chat) error
	GetChatsByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.Chat, error)
	GetChatByID(ctx context.Context, chatID uuid.UUID) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
	CreateMessage(ctx context.Context, msg *models.Message) error
	GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*models.Message, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateChat(ctx context.Context, chat *models.Chat) error {
	query := `
		INSERT INTO chats (id, workspace_id, user_id, title, is_pinned, model_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	chat.ID = uuid.New()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	if chat.Title == "" {
		chat.Title = "New Conversation"
	}
	if chat.ModelName == "" {
		chat.ModelName = "gpt-4o"
	}

	_, err := r.pool.Exec(ctx, query, chat.ID, chat.WorkspaceID, chat.UserID, chat.Title, chat.IsPinned, chat.ModelName, chat.CreatedAt, chat.UpdatedAt)
	return err
}

func (r *repository) GetChatsByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.Chat, error) {
	query := `
		SELECT id, workspace_id, user_id, title, is_pinned, model_name, created_at, updated_at
		FROM chats WHERE workspace_id = $1 AND user_id = $2
		ORDER BY is_pinned DESC, updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*models.Chat
	for rows.Next() {
		var c models.Chat
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.UserID, &c.Title, &c.IsPinned, &c.ModelName, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &c)
	}
	return chats, nil
}

func (r *repository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*models.Chat, error) {
	query := `SELECT id, workspace_id, user_id, title, is_pinned, model_name, created_at, updated_at FROM chats WHERE id = $1`
	var c models.Chat
	err := r.pool.QueryRow(ctx, query, chatID).Scan(&c.ID, &c.WorkspaceID, &c.UserID, &c.Title, &c.IsPinned, &c.ModelName, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repository) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	query := `DELETE FROM chats WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, chatID)
	return err
}

func (r *repository) CreateMessage(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, chat_id, sender_role, content, citations, token_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, msg.ID, msg.ChatID, msg.SenderRole, msg.Content, msg.Citations, msg.TokenCount, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Touch chat updated_at
	_, _ = r.pool.Exec(ctx, `UPDATE chats SET updated_at = NOW() WHERE id = $1`, msg.ChatID)
	return nil
}

func (r *repository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*models.Message, error) {
	query := `
		SELECT id, chat_id, sender_role, content, citations, token_count, created_at
		FROM messages WHERE chat_id = $1 ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderRole, &m.Content, &m.Citations, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, nil
}
