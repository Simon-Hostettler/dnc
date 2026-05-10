package repository

import (
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/util"
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
	Features     []models.FeatureTO
	Notes        []models.NoteTO
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

func (c *CharacterAggregate) AddEmptyFeature() uuid.UUID {
	feat := models.FeatureTO{ID: uuid.New()}
	c.Features = append(c.Features, feat)
	return feat.ID
}

func (c *CharacterAggregate) AddEmptyNote() uuid.UUID {
	note := models.NoteTO{ID: uuid.New()}
	c.Notes = append(c.Notes, note)
	return note.ID
}

func (c *CharacterAggregate) DeleteAttack(id uuid.UUID) {
	c.Attacks = util.Filter(c.Attacks, func(a models.AttackTO) bool {
		return a.ID != id
	})
}

func (c *CharacterAggregate) DeleteItem(id uuid.UUID) {
	c.Items = util.Filter(c.Items, func(i models.ItemTO) bool {
		return i.ID != id
	})
}

func (c *CharacterAggregate) DeleteSpell(id uuid.UUID) {
	c.Spells = util.Filter(c.Spells, func(s models.SpellTO) bool {
		return s.ID != id
	})
}

func (c *CharacterAggregate) DeleteFeature(id uuid.UUID) {
	c.Features = util.Filter(c.Features, func(f models.FeatureTO) bool {
		return f.ID != id
	})
}

func (c *CharacterAggregate) DeleteNote(id uuid.UUID) {
	c.Notes = util.Filter(c.Notes, func(n models.NoteTO) bool {
		return n.ID != id
	})
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
