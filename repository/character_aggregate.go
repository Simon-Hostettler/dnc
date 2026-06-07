package repository

import (
	"fmt"

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

	// shadow is the last known persisted state. Set by the repository on
	// load/create and refreshed after each successful Update. Compared
	// section-by-section in Update so we only write tables that changed.
	shadow *CharacterAggregate
}

// Clone returns a deep copy suitable for use as a shadow. The clone's own
// shadow field is left nil — shadows do not nest.
func (c *CharacterAggregate) Clone() *CharacterAggregate {
	if c == nil {
		return nil
	}
	cp := &CharacterAggregate{}
	if c.Character != nil {
		ch := *c.Character
		ch.SpellSlots = append(models.IntList(nil), c.Character.SpellSlots...)
		ch.SpellSlotsUsed = append(models.IntList(nil), c.Character.SpellSlotsUsed...)
		cp.Character = &ch
	}
	if c.Abilities != nil {
		a := *c.Abilities
		cp.Abilities = &a
	}
	if c.SavingThrows != nil {
		s := *c.SavingThrows
		cp.SavingThrows = &s
	}
	if c.Wallet != nil {
		w := *c.Wallet
		cp.Wallet = &w
	}
	cp.Items = append([]models.ItemTO(nil), c.Items...)
	cp.Spells = append([]models.SpellTO(nil), c.Spells...)
	cp.Attacks = append([]models.AttackTO(nil), c.Attacks...)
	cp.Skills = append([]models.CharacterSkillDetailTO(nil), c.Skills...)
	cp.Features = append([]models.FeatureTO(nil), c.Features...)
	cp.Notes = append([]models.NoteTO(nil), c.Notes...)
	return cp
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

func (c *CharacterAggregate) LongRest() {
	ch := c.Character
	ch.CurrHitPoints = ch.MaxHitPoints
	ch.DeathSaveSuccesses = 0
	ch.DeathSaveFailures = 0
	for i := range ch.SpellSlotsUsed {
		ch.SpellSlotsUsed[i] = 0
	}
}

func (c *CharacterAggregate) Heal(amount int) {
	ch := c.Character
	ch.CurrHitPoints = min(ch.CurrHitPoints+amount, ch.MaxHitPoints)
}

func (c *CharacterAggregate) TakeDamage(amount int) {
	ch := c.Character
	spill := max(0, -(ch.TempHitPoints - amount))
	ch.TempHitPoints = max(ch.TempHitPoints-amount, 0)
	ch.CurrHitPoints = max(ch.CurrHitPoints-spill, 0)
}

func (c *CharacterAggregate) SetTempHP(amount int) {
	c.Character.TempHitPoints = amount
}

func (c *CharacterAggregate) CastSpell(level int) error {
	ch := c.Character
	if level >= len(ch.SpellSlots) || ch.SpellSlots[level] <= 0 {
		return fmt.Errorf("no spell slots at level %d", level)
	}
	if ch.SpellSlotsUsed[level] >= ch.SpellSlots[level] {
		return fmt.Errorf("no available slots at level %d", level)
	}
	ch.SpellSlotsUsed[level]++
	return nil
}
