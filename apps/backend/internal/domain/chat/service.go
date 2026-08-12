package chat

import (
	"context"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type CreateChatRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title       string    `json:"title"`
	ModelName   string    `json:"model_name"`
}

type Service interface {
	CreateChat(ctx context.Context, userID uuid.UUID, req CreateChatRequest) (*models.Chat, error)
	GetWorkspaceChats(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.Chat, error)
	GetChatDetails(ctx context.Context, chatID uuid.UUID) (*models.Chat, []*models.Message, error)
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
	SaveMessage(ctx context.Context, msg *models.Message) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateChat(ctx context.Context, userID uuid.UUID, req CreateChatRequest) (*models.Chat, error) {
	chat := &models.Chat{
		WorkspaceID: req.WorkspaceID,
		UserID:      userID,
		Title:       req.Title,
		ModelName:   req.ModelName,
	}
	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *service) GetWorkspaceChats(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.Chat, error) {
	return s.repo.GetChatsByWorkspace(ctx, workspaceID, userID)
}

func (s *service) GetChatDetails(ctx context.Context, chatID uuid.UUID) (*models.Chat, []*models.Message, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}
	msgs, err := s.repo.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}
	return chat, msgs, nil
}

func (s *service) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	return s.repo.DeleteChat(ctx, chatID)
}

func (s *service) SaveMessage(ctx context.Context, msg *models.Message) error {
	return s.repo.CreateMessage(ctx, msg)
}
