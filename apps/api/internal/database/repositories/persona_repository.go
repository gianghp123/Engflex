package repositories

import (
	"context"

	"gorm.io/gorm"

	"engflex-api/internal/database/models"
)

// PersonaRepository is the DB access layer for AI conversation partners.
type PersonaRepository interface {
	List(ctx context.Context, limit, offset int) ([]*models.Persona, error)
	Count(ctx context.Context) (int64, error)
	GetByID(ctx context.Context, id string) (*models.Persona, error)
	Create(ctx context.Context, p *models.Persona) error
	Update(ctx context.Context, p *models.Persona) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) PersonaRepository
}

type personaRepository struct {
	db *gorm.DB
}

// NewPersonaRepository builds the repository over the shared DB handle.
func NewPersonaRepository(db *gorm.DB) PersonaRepository {
	return &personaRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *personaRepository) WithTx(tx *gorm.DB) PersonaRepository {
	return &personaRepository{db: tx}
}

func (r *personaRepository) List(ctx context.Context, limit, offset int) ([]*models.Persona, error) {
	var out []*models.Persona
	if err := r.db.WithContext(ctx).
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *personaRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.Persona{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *personaRepository) GetByID(ctx context.Context, id string) (*models.Persona, error) {
	if _, err := parseID(id, "personaId"); err != nil {
		return nil, err
	}
	var m models.Persona
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *personaRepository) Create(ctx context.Context, p *models.Persona) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *personaRepository) Update(ctx context.Context, p *models.Persona) error {
	result := r.db.WithContext(ctx).Save(p)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *personaRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "personaId"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.Persona{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
