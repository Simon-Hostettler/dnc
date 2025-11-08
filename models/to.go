package models

import (
	"time"

	"github.com/google/uuid"
)

// CharacterTO maps directly to the `character` table.
type CharacterTO struct {
	ID                  uuid.UUID `db:"id"`
	Name                string    `db:"name"`
	ClassLevels         string    `db:"class_levels"`
	Race                string    `db:"race"`
	Alignment           string    `db:"alignment"`
	ProficiencyBonus    int       `db:"proficiency_bonus"`
	ArmorClass          int       `db:"armor_class"`
	Initiative          int       `db:"initiative"`
	Speed               int       `db:"speed"`
	MaxHitPoints        int       `db:"max_hit_points"`
	CurrHitPoints       int       `db:"curr_hit_points"`
	TempHitPoints       int       `db:"temp_hit_points"`
	HitDice             string    `db:"hit_dice"`
	UsedHitDice         string    `db:"used_hit_dice"`
	DeathSaveSuccesses  int       `db:"death_save_successes"`
	DeathSaveFailures   int       `db:"death_save_failures"`
	Actions             string    `db:"actions"`
	BonusActions        string    `db:"bonus_actions"`
	SpellSlots          IntList   `db:"spell_slots"`
	SpellSlotsUsed      IntList   `db:"spell_slots_used"`
	SpellcastingAbility string    `db:"spellcasting_ability"`
	SpellSaveDC         int       `db:"spell_save_dc"`
	SpellAttackBonus    int       `db:"spell_attack_bonus"`
	Age                 int       `db:"age"`
	Height              string    `db:"height"`
	Weight              string    `db:"weight"`
	Eyes                string    `db:"eyes"`
	Skin                string    `db:"skin"`
	Hair                string    `db:"hair"`
	Appearance          string    `db:"appearance"`
	Backstory           string    `db:"backstory"`
	Personality         string    `db:"personality"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}

// ItemTO maps to the `item` table.
type ItemTO struct {
	ID              uuid.UUID  `db:"id"`
	CharacterID     uuid.UUID  `db:"character_id"`
	Name            string     `db:"name"`
	Equipped        Equippable `db:"equipped"`
	AttunementSlots int        `db:"attunement_slots"`
	Quantity        int        `db:"quantity"`
	Description     string     `db:"description"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

// WalletTO maps to the `wallet` table.
type WalletTO struct {
	CharacterID uuid.UUID `db:"character_id"`
	Copper      int       `db:"copper"`
	Silver      int       `db:"silver"`
	Electrum    int       `db:"electrum"`
	Gold        int       `db:"gold"`
	Platinum    int       `db:"platinum"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// SpellTO maps to the `spell` table.
type SpellTO struct {
	ID          uuid.UUID `db:"id"`
	CharacterID uuid.UUID `db:"character_id"`
	Name        string    `db:"name"`
	Level       int       `db:"level"`
	// Prepared is stored as 0/1 integer in DB for compactness.
	Prepared    int       `db:"prepared"`
	Damage      string    `db:"damage"`
	CastingTime string    `db:"casting_time"`
	Range       string    `db:"range"`
	Duration    string    `db:"duration"`
	Components  string    `db:"components"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// AttackTO maps to the `attacks` table.
type AttackTO struct {
	ID          uuid.UUID `db:"id"`
	CharacterID uuid.UUID `db:"character_id"`
	Name        string    `db:"name"`
	Bonus       int       `db:"bonus"`
	Damage      string    `db:"damage"`
	DamageType  string    `db:"damage_type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// AbilitiesTO maps to the `abilities` table.
type AbilitiesTO struct {
	CharacterID  uuid.UUID `db:"character_id"`
	Strength     int       `db:"strength"`
	Dexterity    int       `db:"dexterity"`
	Constitution int       `db:"constitution"`
	Intelligence int       `db:"intelligence"`
	Wisdom       int       `db:"wisdom"`
	Charisma     int       `db:"charisma"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// SavingThrowsTO maps to the `saving_throws` table.
type SavingThrowsTO struct {
	CharacterID             uuid.UUID `db:"character_id"`
	StrengthProficiency     int       `db:"strength_proficiency"`
	DexterityProficiency    int       `db:"dexterity_proficiency"`
	ConstitutionProficiency int       `db:"constitution_proficiency"`
	IntelligenceProficiency int       `db:"intelligence_proficiency"`
	WisdomProficiency       int       `db:"wisdom_proficiency"`
	CharismaProficiency     int       `db:"charisma_proficiency"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
}

// SkillDefinitionTO maps to the canonical `skill_definition` table.
type SkillDefinitionTO struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	Ability string `db:"ability"`
}

// CharacterSkillTO maps to the `character_skill` table.
type CharacterSkillTO struct {
	ID             uuid.UUID `db:"id"`
	CharacterID    uuid.UUID `db:"character_id"`
	SkillID        int       `db:"skill_id"`
	Proficiency    int       `db:"proficiency"`
	CustomModifier int       `db:"custom_modifier"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CharacterSkillDetailTO represents a joined view of character_skill with skill_definition
type CharacterSkillDetailTO struct {
	ID             uuid.UUID `db:"id"`
	CharacterID    uuid.UUID `db:"character_id"`
	SkillID        int       `db:"skill_id"`
	Proficiency    int       `db:"proficiency"`
	CustomModifier int       `db:"custom_modifier"`
	SkillName      string    `db:"skill_name"`
	SkillAbility   string    `db:"skill_ability"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// CharacterSkillTO maps to the `features` table.
type FeatureTO struct {
	ID          uuid.UUID `db:"id"`
	CharacterID uuid.UUID `db:"character_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
