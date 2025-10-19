package repository

import (
	"context"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

// CharacterAggregate represents a character and all of its dependent rows.
type CharacterAggregate struct {
	Character *models.CharacterTO
	Abilities *models.AbilitiesTO
	Wallet    *models.WalletTO
	Items     []models.ItemTO
	Spells    []models.SpellTO
	Attacks   []models.AttackTO
	Skills    []models.CharacterSkillDetailTO
}

// CharacterRepository defines core operations for loading and persisting characters.
type CharacterRepository interface {
	Create(ctx context.Context, agg *CharacterAggregate) (uuid.UUID, error)
	CreateEmpty(ctx context.Context, name string) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*CharacterAggregate, error)
	ListSummary(ctx context.Context) ([]models.CharacterSummary, error)
	Delete(ctx context.Context, id uuid.UUID) error

	UpdateCharacter(ctx context.Context, c models.CharacterTO) error
	UpdateCharacterFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error

	ListSkillDefinitions(ctx context.Context) ([]models.SkillDefinitionTO, error)

	AddSpell(ctx context.Context, characterID uuid.UUID, sp models.SpellTO) (uuid.UUID, error)
	DeleteSpell(ctx context.Context, spellID uuid.UUID) error
	ListSpellsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.SpellTO, error)
	UpdateSpellFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error

	AddItem(ctx context.Context, characterID uuid.UUID, it models.ItemTO) (uuid.UUID, error)
	DeleteItem(ctx context.Context, itemID uuid.UUID) error
	ListItemsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.ItemTO, error)
	UpdateItemFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error

	GetWallet(ctx context.Context, characterID uuid.UUID) (*models.WalletTO, error)
	UpdateWalletFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error

	AddAttack(ctx context.Context, characterID uuid.UUID, a models.AttackTO) (uuid.UUID, error)
	DeleteAttack(ctx context.Context, attackID uuid.UUID) error
	ListAttacksByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.AttackTO, error)

	UpsertSkill(ctx context.Context, characterID uuid.UUID, skillID int, proficiency int, customModifier int) error
	DeleteSkill(ctx context.Context, characterID uuid.UUID, skillID int) error
	ListSkillsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillTO, error)
	ListSkillDetailsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillDetailTO, error)

	GetAbilities(ctx context.Context, characterID uuid.UUID) (*models.AbilitiesTO, error)
	UpsertAbilities(ctx context.Context, characterID uuid.UUID, ab models.AbilitiesTO) error
}
