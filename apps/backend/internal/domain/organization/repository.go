package organization

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopor-ai/lopor/internal/domain/models"
)

type OrganizationMember struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

type Repository interface {
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	AddOrganizationMember(ctx context.Context, orgID, userID uuid.UUID, role string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, logo_url, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	org.ID = uuid.New()
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query, org.ID, org.Name, org.Slug, org.LogoURL, org.OwnerID, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *repository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `SELECT id, name, slug, logo_url, owner_id, created_at, updated_at FROM organizations WHERE id = $1`
	var org models.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Slug, &org.LogoURL, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	query := `
		SELECT id, name, slug, logo_url, owner_id, created_at, updated_at
		FROM organizations WHERE owner_id = $1 ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.LogoURL, &o.OwnerID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, &o)
	}
	return orgs, nil
}

func (r *repository) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error) {
	query := `
		SELECT o.id, u.id, u.email, u.full_name, u.role, o.created_at
		FROM users u
		JOIN organizations o ON o.owner_id = u.id
		WHERE o.id = $1
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*OrganizationMember
	for rows.Next() {
		var m OrganizationMember
		if err := rows.Scan(&m.OrganizationID, &m.UserID, &m.Email, &m.FullName, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *repository) AddOrganizationMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	return nil
}
