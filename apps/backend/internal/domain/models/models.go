package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a system user entity
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	FullName      string    `json:"full_name" db:"full_name"`
	AvatarURL     *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	Role          string    `json:"role" db:"role"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Profile represents user preferences and biographical info
type Profile struct {
	UserID         uuid.UUID       `json:"user_id" db:"user_id"`
	Bio            *string         `json:"bio,omitempty" db:"bio"`
	JobTitle       *string         `json:"job_title,omitempty" db:"job_title"`
	Department     *string         `json:"department,omitempty" db:"department"`
	PreferredTheme string          `json:"preferred_theme" db:"preferred_theme"`
	Language       string          `json:"language" db:"language"`
	Preferences    json.RawMessage `json:"preferences" db:"preferences"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a user authentication session token
type RefreshToken struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	UserID     uuid.UUID       `json:"user_id" db:"user_id"`
	TokenHash  string          `json:"-" db:"token_hash"`
	DeviceInfo json.RawMessage `json:"device_info,omitempty" db:"device_info"`
	IPAddress  string          `json:"ip_address" db:"ip_address"`
	IsRevoked  bool            `json:"is_revoked" db:"is_revoked"`
	ExpiresAt  time.Time       `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// Organization represents an enterprise tenant account
type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	LogoURL   *string   `json:"logo_url,omitempty" db:"logo_url"`
	OwnerID   uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Workspace represents a project boundary within an organization
type Workspace struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Slug           string    `json:"slug" db:"slug"`
	Description    *string   `json:"description,omitempty" db:"description"`
	Icon           *string   `json:"icon,omitempty" db:"icon"`
	IsPublic       bool      `json:"is_public" db:"is_public"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// WorkspaceMember represents membership and permissions in a workspace
type WorkspaceMember struct {
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// Folder represents hierarchical document categories
type Folder struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Name        string     `json:"name" db:"name"`
	Color       *string    `json:"color,omitempty" db:"color"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Document represents rich text and markdown content
type Document struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	WorkspaceID  uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	FolderID     *uuid.UUID      `json:"folder_id,omitempty" db:"folder_id"`
	Title        string          `json:"title" db:"title"`
	Content      *string         `json:"content,omitempty" db:"content"`
	ContentJSON  json.RawMessage `json:"content_json,omitempty" db:"content_json"`
	Icon         *string         `json:"icon,omitempty" db:"icon"`
	CoverImage   *string         `json:"cover_image,omitempty" db:"cover_image"`
	IsArchived   bool            `json:"is_archived" db:"is_archived"`
	IsPublished  bool            `json:"is_published" db:"is_published"`
	CreatedBy    uuid.UUID       `json:"created_by" db:"created_by"`
	LastEditedBy uuid.UUID       `json:"last_edited_by" db:"last_edited_by"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// Chat represents an AI conversation context
type Chat struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	IsPinned    bool      `json:"is_pinned" db:"is_pinned"`
	ModelName   string    `json:"model_name" db:"model_name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Message represents an individual message within an AI Chat session
type Message struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	ChatID     uuid.UUID       `json:"chat_id" db:"chat_id"`
	SenderRole string          `json:"sender_role" db:"sender_role"`
	Content    string          `json:"content" db:"content"`
	Citations  json.RawMessage `json:"citations,omitempty" db:"citations"`
	TokenCount int             `json:"token_count" db:"token_count"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// Embedding represents a chunk vector stored in pgvector
type Embedding struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id" db:"workspace_id"`
	DocumentID  *uuid.UUID      `json:"document_id,omitempty" db:"document_id"`
	FileID      *uuid.UUID      `json:"file_id,omitempty" db:"file_id"`
	ChunkIndex  int             `json:"chunk_index" db:"chunk_index"`
	ChunkText   string          `json:"chunk_text" db:"chunk_text"`
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}
