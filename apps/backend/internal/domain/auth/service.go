package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
	"github.com/lopor-ai/lopor/pkg/jwt"
	"github.com/lopor-ai/lopor/pkg/password"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest, ip string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshTokenString string) error
	GetMe(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, user.Role, s.jwtSecret, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	rawRefreshToken := uuid.New().String()
	hash := hashToken(rawRefreshToken)

	rfToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.CreateRefreshToken(ctx, rfToken); err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		User:         user,
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest, ip string) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !password.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, user.Role, s.jwtSecret, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	rawRefreshToken := uuid.New().String()
	hash := hashToken(rawRefreshToken)

	rfToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		IPAddress: ip,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.repo.CreateRefreshToken(ctx, rfToken); err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		User:         user,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshTokenString string) (*AuthResponse, error) {
	hash := hashToken(refreshTokenString)
	token, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil || token == nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.GetUserByID(ctx, token.UserID)
	if err != nil || user == nil {
		return nil, errors.New("associated user not found")
	}

	_ = s.repo.RevokeRefreshToken(ctx, hash)

	newAccessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, user.Role, s.jwtSecret, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	newRawRefreshToken := uuid.New().String()
	newHash := hashToken(newRawRefreshToken)

	rfToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	_ = s.repo.CreateRefreshToken(ctx, rfToken)

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
		User:         user,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshTokenString string) error {
	hash := hashToken(refreshTokenString)
	return s.repo.RevokeRefreshToken(ctx, hash)
}

func (s *service) GetMe(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
