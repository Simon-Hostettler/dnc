package repository

import (
	"context"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

// CharacterAggregate represents a character and all of its dependent rows.
type CharacterAggregate struct {
	Character    *models.CharacterTO
	Abilities    *models.AbilitiesTO
	SavingThrows *models.SavingThrowsTO
	Wallet       *models.WalletTO
	Items        []models.ItemTO
	Spells       []models.SpellTO
	Attacks      []models.AttackTO
	Skills       []models.CharacterSkillDetailTO
}

func (c *CharacterAggregate) AddEmptyItem() uuid.UUID {
	item := models.ItemTO{ID: uuid.New()}
	c.Items = append(c.Items, item)
	return item.ID
}

func (c *CharacterAggregate) AddEmptySpell(l int) uuid.UUID {
	spell := models.SpellTO{ID: uuid.New(), Level: l}
	c.Spells = append(c.Spells, spell)
	return spell.ID
}

func (c *CharacterAggregate) DeleteItem(id uuid.UUID) {
	newSpells := []models.SpellTO{}
	for i := range c.Spells {
		if c.Spells[i].ID != id {
			newSpells = append(newSpells, c.Spells[i])
		}
	}
	c.Spells = newSpells
}

// CharacterRepository defines core operations for loading and persisting characters.
type CharacterRepository interface {
	CreateEmpty(ctx context.Context, name string) (uuid.UUID, error)
	Update(ctx context.Context, c *CharacterAggregate) error
	GetByID(ctx context.Context, id uuid.UUID) (*CharacterAggregate, error)
	ListSummary(ctx context.Context) ([]models.CharacterSummary, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
