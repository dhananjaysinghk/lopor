package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, hash string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateUser(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, avatar_url, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	u.ID = uuid.New()
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.Role == "" {
		u.Role = "user"
	}

	_, err := r.pool.Exec(ctx, query, u.ID, u.Email, u.PasswordHash, u.FullName, u.AvatarURL, u.Role, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}
	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, email_verified, is_active, role, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = true
	`
	row := r.pool.QueryRow(ctx, query, email)

	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarURL, &u.EmailVerified, &u.IsActive, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, email_verified, is_active, role, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = true
	`
	row := r.pool.QueryRow(ctx, query, id)

	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarURL, &u.EmailVerified, &u.IsActive, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.DeviceInfo, token.IPAddress, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *repository) GetRefreshToken(ctx context.Context, hash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, is_revoked, expires_at, created_at
		FROM refresh_tokens WHERE token_hash = $1 AND is_revoked = false AND expires_at > NOW()
	`
	var t models.RefreshToken
	err := r.pool.QueryRow(ctx, query, hash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.DeviceInfo, &t.IPAddress, &t.IsRevoked, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, hash string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, hash)
	return err
}
