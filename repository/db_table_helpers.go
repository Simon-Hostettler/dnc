package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/models"
)

func nonZeroOr(t, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t
}

// childTable describes a 1:N table keyed by its own id PK with a character_id FK.
type childTable[T any] struct {
	name    string
	columns []string
	orderBy string
	values  func(item *T, charID uuid.UUID) []any
}

// ownedTable describes a 1:1 table whose primary key is character_id alone.
type ownedTable[T any] struct {
	name    string
	columns []string
	values  func(item T, charID uuid.UUID) []any
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

func insertStmt(table string, columns []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "), placeholders(len(columns)))
}

func deleteByCharacter(ctx context.Context, tx *sqlx.Tx, table string, charID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE character_id=?", table), charID)
	return err
}

func replaceAll[T any](ctx context.Context, tx *sqlx.Tx, spec childTable[T], charID uuid.UUID, items []T) error {
	if err := deleteByCharacter(ctx, tx, spec.name, charID); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	stmt := insertStmt(spec.name, spec.columns)
	for i := range items {
		if _, err := tx.ExecContext(ctx, stmt, spec.values(&items[i], charID)...); err != nil {
			return err
		}
	}
	return nil
}

func upsertOne[T any](ctx context.Context, tx *sqlx.Tx, spec ownedTable[T], charID uuid.UUID, item *T) error {
	if item == nil {
		return nil
	}
	if err := deleteByCharacter(ctx, tx, spec.name, charID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, insertStmt(spec.name, spec.columns), spec.values(*item, charID)...)
	return err
}

func selectAll[T any](ctx context.Context, q sqlx.QueryerContext, spec childTable[T], charID uuid.UUID) ([]T, error) {
	var rows []T
	query := fmt.Sprintf("SELECT * FROM %s WHERE character_id=? ORDER BY %s", spec.name, spec.orderBy)
	if err := sqlx.SelectContext(ctx, q, &rows, query, charID); err != nil {
		return nil, err
	}
	return rows, nil
}

func getOne[T any](ctx context.Context, q sqlx.QueryerContext, table string, charID uuid.UUID) (*T, error) {
	var v T
	query := fmt.Sprintf("SELECT * FROM %s WHERE character_id=?", table)
	if err := sqlx.GetContext(ctx, q, &v, query, charID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

var itemTable = childTable[models.ItemTO]{
	name:    "item",
	columns: []string{"id", "character_id", "name", "is_equippable", "equipped", "attunement_slots", "quantity", "description", "created_at", "updated_at"},
	orderBy: "name ASC",
	values: func(it *models.ItemTO, charID uuid.UUID) []any {
		if it.ID == uuid.Nil {
			it.ID = uuid.New()
		}
		now := time.Now()
		return []any{it.ID, charID, it.Name, it.IsEquippable, it.Equipped, it.AttunementSlots, it.Quantity, it.Description, nonZeroOr(it.CreatedAt, now), now}
	},
}

var spellTable = childTable[models.SpellTO]{
	name:    "spell",
	columns: []string{"id", "character_id", "name", "school", "level", "prepared", "concentration", "ritual", "spell_source", "damage", "casting_time", "range", "duration", "components", "description", "created_at", "updated_at"},
	orderBy: "level ASC, name ASC",
	values: func(s *models.SpellTO, charID uuid.UUID) []any {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		now := time.Now()
		return []any{s.ID, charID, s.Name, s.School, s.Level, s.Prepared, s.Concentration, s.Ritual, s.SpellSource, s.Damage, s.CastingTime, s.Range, s.Duration, s.Components, s.Description, nonZeroOr(s.CreatedAt, now), now}
	},
}

var attackTable = childTable[models.AttackTO]{
	name:    "attacks",
	columns: []string{"id", "character_id", "name", "bonus", "damage", "damage_type", "created_at", "updated_at"},
	orderBy: "created_at ASC",
	values: func(a *models.AttackTO, charID uuid.UUID) []any {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		now := time.Now()
		return []any{a.ID, charID, a.Name, a.Bonus, a.Damage, a.DamageType, nonZeroOr(a.CreatedAt, now), now}
	},
}

var featureTable = childTable[models.FeatureTO]{
	name:    "features",
	columns: []string{"id", "character_id", "name", "description", "created_at", "updated_at"},
	orderBy: "name ASC",
	values: func(f *models.FeatureTO, charID uuid.UUID) []any {
		if f.ID == uuid.Nil {
			f.ID = uuid.New()
		}
		now := time.Now()
		return []any{f.ID, charID, f.Name, f.Description, nonZeroOr(f.CreatedAt, now), now}
	},
}

var noteTable = childTable[models.NoteTO]{
	name:    "notes",
	columns: []string{"id", "character_id", "title", "note", "created_at", "updated_at"},
	orderBy: "title ASC",
	values: func(n *models.NoteTO, charID uuid.UUID) []any {
		if n.ID == uuid.Nil {
			n.ID = uuid.New()
		}
		now := time.Now()
		return []any{n.ID, charID, n.Title, n.Note, nonZeroOr(n.CreatedAt, now), now}
	},
}

var skillTable = childTable[models.CharacterSkillTO]{
	name:    "character_skill",
	columns: []string{"id", "character_id", "skill_id", "proficiency", "custom_modifier", "created_at", "updated_at"},
	orderBy: "skill_id ASC",
	values: func(s *models.CharacterSkillTO, charID uuid.UUID) []any {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		now := time.Now()
		return []any{s.ID, charID, s.SkillID, s.Proficiency, s.CustomModifier, nonZeroOr(s.CreatedAt, now), now}
	},
}

var abilitiesTable = ownedTable[models.AbilitiesTO]{
	name:    "abilities",
	columns: []string{"character_id", "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma", "created_at", "updated_at"},
	values: func(a models.AbilitiesTO, charID uuid.UUID) []any {
		now := time.Now()
		return []any{charID, a.Strength, a.Dexterity, a.Constitution, a.Intelligence, a.Wisdom, a.Charisma, nonZeroOr(a.CreatedAt, now), now}
	},
}

var savingThrowsTable = ownedTable[models.SavingThrowsTO]{
	name:    "saving_throws",
	columns: []string{"character_id", "strength_proficiency", "dexterity_proficiency", "constitution_proficiency", "intelligence_proficiency", "wisdom_proficiency", "charisma_proficiency", "created_at", "updated_at"},
	values: func(s models.SavingThrowsTO, charID uuid.UUID) []any {
		now := time.Now()
		return []any{charID, s.StrengthProficiency, s.DexterityProficiency, s.ConstitutionProficiency, s.IntelligenceProficiency, s.WisdomProficiency, s.CharismaProficiency, nonZeroOr(s.CreatedAt, now), now}
	},
}

var walletTable = ownedTable[models.WalletTO]{
	name:    "wallet",
	columns: []string{"character_id", "copper", "silver", "electrum", "gold", "platinum", "created_at", "updated_at"},
	values: func(w models.WalletTO, charID uuid.UUID) []any {
		now := time.Now()
		return []any{charID, w.Copper, w.Silver, w.Electrum, w.Gold, w.Platinum, nonZeroOr(w.CreatedAt, now), now}
	},
}
