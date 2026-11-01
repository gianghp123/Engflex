package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/scenarios/dtos/requests"
	"engflex-api/internal/utils"
)

// PersonaService owns the persona catalog.
type PersonaService struct {
	personas repositories.PersonaRepository
}

// NewPersonaService builds the service over the repository interface.
func NewPersonaService(personas repositories.PersonaRepository) *PersonaService {
	return &PersonaService{personas: personas}
}

// List pages the persona catalog.
func (s *PersonaService) List(ctx context.Context, req requests.ListPersonas) ([]*models.Persona, int64, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	offset := utils.Offset(page, pageSize)

	items, err := s.personas.List(ctx, pageSize, offset)
	if err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "list personas failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.personas.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "count personas failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "personas listed", "count", len(items), "total", total)
	return items, total, nil
}

func (s *PersonaService) GetByID(ctx context.Context, id string) (*models.Persona, error) {
	m, err := s.personas.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "get persona failed", appErr, "personaID", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "persona found", "personaID", m.ID)
	return m, nil
}

// Create/Update/Delete manage AI conversation partners (admin-only,
// route-gated). Personas have no draft concept — visible immediately.
func (s *PersonaService) Create(ctx context.Context, req requests.CreatePersona) (*models.Persona, error) {
	m := &models.Persona{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create persona failed", err)
		return nil, common.Internal()
	}
	if err := s.personas.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "create persona failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "persona created", "personaID", m.ID)
	return m, nil
}

func (s *PersonaService) Update(ctx context.Context, id string, req requests.UpdatePersona) (*models.Persona, error) {
	m, err := s.personas.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "get persona for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update persona failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.personas.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "update persona failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "persona updated", "personaID", id)
	return m, nil
}

func (s *PersonaService) Delete(ctx context.Context, id string) error {
	if err := s.personas.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "persona")
		logger.Report(ctx, "delete persona failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "persona deleted", "id", id)
	return nil
}
