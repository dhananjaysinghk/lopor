package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
	"github.com/lopor-ai/lopor/pkg/email"
)

type CreateOrgRequest struct {
	Name    string  `json:"name"`
	LogoURL *string `json:"logo_url,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Service interface {
	CreateOrganization(ctx context.Context, ownerID uuid.UUID, req CreateOrgRequest) (*models.Organization, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	InviteMember(ctx context.Context, orgID uuid.UUID, req InviteMemberRequest) error
}

type service struct {
	repo   Repository
	mailer *email.Mailer
}

func NewService(repo Repository, mailer *email.Mailer) Service {
	return &service{
		repo:   repo,
		mailer: mailer,
	}
}

func (s *service) CreateOrganization(ctx context.Context, ownerID uuid.UUID, req CreateOrgRequest) (*models.Organization, error) {
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	org := &models.Organization{
		Name:    req.Name,
		Slug:    slug,
		LogoURL: req.LogoURL,
		OwnerID: ownerID,
	}
	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *service) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	return s.repo.GetUserOrganizations(ctx, userID)
}

func (s *service) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error) {
	return s.repo.GetOrganizationMembers(ctx, orgID)
}

func (s *service) InviteMember(ctx context.Context, orgID uuid.UUID, req InviteMemberRequest) error {
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return err
	}

	token := uuid.New().String()
	inviteLink := fmt.Sprintf("http://localhost:3000/invite/%s", token)

	return s.mailer.SendInvitationEmail(req.Email, org.Name, inviteLink)
}
