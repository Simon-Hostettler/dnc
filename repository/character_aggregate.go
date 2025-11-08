package repository

import (
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
	Features     []models.FeatureTO
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

func (c *CharacterAggregate) DeleteAttack(id uuid.UUID) {
	newAttacks := []models.AttackTO{}
	for i := range c.Attacks {
		if c.Attacks[i].ID != id {
			newAttacks = append(newAttacks, c.Attacks[i])
		}
	}
	c.Attacks = newAttacks
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

func (c *CharacterAggregate) DeleteFeature(id uuid.UUID) {
	newFeatures := []models.FeatureTO{}
	for i := range c.Features {
		if c.Features[i].ID != id {
			newFeatures = append(newFeatures, c.Features[i])
		}
	}
	c.Features = newFeatures
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
