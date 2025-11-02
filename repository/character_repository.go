package repository

import (
	"context"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

// CharacterRepository defines core operations for loading and persisting characters.
type CharacterRepository interface {
	CreateEmpty(ctx context.Context, name string) (uuid.UUID, error)
	Update(ctx context.Context, c *CharacterAggregate) error
	GetByID(ctx context.Context, id uuid.UUID) (*CharacterAggregate, error)
	ListSummary(ctx context.Context) ([]models.CharacterSummary, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
