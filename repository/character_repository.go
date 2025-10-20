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

// Helper methods - Modify TOs not database, changes have to be written back (See command.WriteBackRequest)

func (c *CharacterAggregate) AddEmptyItem() uuid.UUID {
	item := models.ItemTO{ID: uuid.New()}
	c.Items = append(c.Items, item)
	return item.ID
}

func (c *CharacterAggregate) AddEmptyAttack() uuid.UUID {
	attack := models.AttackTO{ID: uuid.New()}
	c.Attacks = append(c.Attacks, attack)
	return attack.ID
}

func (c *CharacterAggregate) AddEmptySpell(l int) uuid.UUID {
	spell := models.SpellTO{ID: uuid.New(), Level: l}
	c.Spells = append(c.Spells, spell)
	return spell.ID
}

func (c *CharacterAggregate) DeleteItem(id uuid.UUID) {
	newItems := []models.ItemTO{}
	for i := range c.Items {
		if c.Items[i].ID != id {
			newItems = append(newItems, c.Items[i])
		}
	}
	c.Items = newItems
}

func (c *CharacterAggregate) DeleteSpell(id uuid.UUID) {
	newSpells := []models.SpellTO{}
	for i := range c.Spells {
		if c.Spells[i].ID != id {
			newSpells = append(newSpells, c.Spells[i])
		}
	}
	c.Spells = newSpells
}

func (c *CharacterAggregate) GetSpellsByLevel(l int) []*models.SpellTO {
	spells := []*models.SpellTO{}
	for i := range c.Spells {
		if c.Spells[i].Level == l {
			spells = append(spells, &c.Spells[i])
		}
	}
	return spells
}
