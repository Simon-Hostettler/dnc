package repository

import (
	"context"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

// CharacterAggregate represents a character and all of its dependent rows.
type CharacterAggregate struct {
	Character models.CharacterTO
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
	GetByName(ctx context.Context, name string) (*CharacterAggregate, error)
	List(ctx context.Context) ([]models.CharacterTO, error)
	ListSummary(ctx context.Context) ([]models.CharacterSummary, error)
	Update(ctx context.Context, agg *CharacterAggregate) error
	Delete(ctx context.Context, id uuid.UUID) error

	UpdateCharacter(ctx context.Context, c models.CharacterTO) error

	ListSkillDefinitions(ctx context.Context) ([]models.SkillDefinitionTO, error)

	AddSpell(ctx context.Context, characterID uuid.UUID, sp models.SpellTO) (uuid.UUID, error)
	UpdateSpell(ctx context.Context, sp models.SpellTO) error
	DeleteSpell(ctx context.Context, spellID uuid.UUID) error
	GetSpell(ctx context.Context, spellID uuid.UUID) (*models.SpellTO, error)
	ListSpellsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.SpellTO, error)

	AddItem(ctx context.Context, characterID uuid.UUID, it models.ItemTO) (uuid.UUID, error)
	UpdateItem(ctx context.Context, it models.ItemTO) error
	DeleteItem(ctx context.Context, itemID uuid.UUID) error
	GetItem(ctx context.Context, itemID uuid.UUID) (*models.ItemTO, error)
	ListItemsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.ItemTO, error)

	AddAttack(ctx context.Context, characterID uuid.UUID, a models.AttackTO) (uuid.UUID, error)
	UpdateAttack(ctx context.Context, a models.AttackTO) error
	DeleteAttack(ctx context.Context, attackID uuid.UUID) error
	GetAttack(ctx context.Context, attackID uuid.UUID) (*models.AttackTO, error)
	ListAttacksByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.AttackTO, error)

	UpsertSkill(ctx context.Context, characterID uuid.UUID, skillID int, proficiency int, customModifier int) error
	DeleteSkill(ctx context.Context, characterID uuid.UUID, skillID int) error
	ListSkillsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillTO, error)
	ListSkillDetailsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillDetailTO, error)

	GetAbilities(ctx context.Context, characterID uuid.UUID) (*models.AbilitiesTO, error)
	UpsertAbilities(ctx context.Context, characterID uuid.UUID, ab models.AbilitiesTO) error
	GetWallet(ctx context.Context, characterID uuid.UUID) (*models.WalletTO, error)
	UpsertWallet(ctx context.Context, characterID uuid.UUID, w models.WalletTO) error
}
