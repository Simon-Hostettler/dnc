package repository

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/util"
)

// DBCharacterRepository is a CharacterRepository backed by sqlx and DuckDB.
type DBCharacterRepository struct {
	db *sqlx.DB
}

func NewDBCharacterRepository(db *sqlx.DB) *DBCharacterRepository {
	return &DBCharacterRepository{db: db}
}

func (r *DBCharacterRepository) Create(ctx context.Context, agg *CharacterAggregate) (uuid.UUID, error) {
	var newID uuid.UUID
	err := r.withTx(ctx, func(tx *sqlx.Tx) error {
		c := agg.Character
		ensureSpellSlots(c)
		query := `
            INSERT INTO character (
                name, class_levels, race, alignment,
                proficiency_bonus, armor_class, initiative, speed,
                max_hit_points, curr_hit_points, temp_hit_points,
                hit_dice, used_hit_dice, death_save_successes, death_save_failures,
				actions, bonus_actions, spell_slots, spell_slots_used,
                spellcasting_ability, spell_save_dc, spell_attack_bonus,
				age, height, weight, eyes, skin, hair, appearance, backstory,
				personality
            ) VALUES (
                ?,?,?,?,?,?,?,?,?,?,?,
                ?,?,?,?,?,?,?,?,?,?,?,?,
				?,?,?,?,?,?,?,?
			) RETURNING id`
		row := tx.QueryRowxContext(ctx, query,
			c.Name, c.ClassLevels, c.Race, c.Alignment,
			c.ProficiencyBonus, c.ArmorClass, c.Initiative, c.Speed,
			c.MaxHitPoints, c.CurrHitPoints, c.TempHitPoints,
			c.HitDice, c.UsedHitDice, c.DeathSaveSuccesses, c.DeathSaveFailures,
			c.Actions, c.BonusActions, c.SpellSlots, c.SpellSlotsUsed,
			c.SpellcastingAbility, c.SpellSaveDC, c.SpellAttackBonus,
			c.Age, c.Height, c.Weight, c.Eyes, c.Skin, c.Hair, c.Appearance, c.Backstory,
			c.Personality,
		)
		if err := row.Scan(&newID); err != nil {
			return err
		}

		if err := upsertOne(ctx, tx, abilitiesTable, newID, agg.Abilities); err != nil {
			return err
		}
		if err := upsertOne(ctx, tx, savingThrowsTable, newID, agg.SavingThrows); err != nil {
			return err
		}
		if err := upsertOne(ctx, tx, walletTable, newID, agg.Wallet); err != nil {
			return err
		}
		if err := replaceAll(ctx, tx, itemTable, newID, agg.Items); err != nil {
			return err
		}
		if err := replaceAll(ctx, tx, spellTable, newID, agg.Spells); err != nil {
			return err
		}
		if err := replaceAll(ctx, tx, attackTable, newID, agg.Attacks); err != nil {
			return err
		}
		if err := replaceAll(ctx, tx, featureTable, newID, agg.Features); err != nil {
			return err
		}
		if err := replaceAll(ctx, tx, noteTable, newID, agg.Notes); err != nil {
			return err
		}
		skills := util.Map(agg.Skills, func(s models.CharacterSkillDetailTO) models.CharacterSkillTO { return s.ToCharacterSkillTO() })
		if err := replaceAll(ctx, tx, skillTable, newID, skills); err != nil {
			return err
		}
		agg.Character.ID = newID
		return nil
	})
	if err == nil {
		agg.shadow = agg.Clone()
	}
	return newID, err
}

func (r *DBCharacterRepository) CreateEmpty(ctx context.Context, name string) (uuid.UUID, error) {
	cSkills := []models.CharacterSkillTO{}
	agg := CharacterAggregate{
		Character:    &models.CharacterTO{Name: name},
		Abilities:    &models.AbilitiesTO{},
		SavingThrows: &models.SavingThrowsTO{},
		Wallet:       &models.WalletTO{},
		Items:        []models.ItemTO{},
		Spells:       []models.SpellTO{},
		Attacks:      []models.AttackTO{},
		Skills:       []models.CharacterSkillDetailTO{},
		Features:     []models.FeatureTO{},
		Notes:        []models.NoteTO{},
	}
	newID, err := r.Create(ctx, &agg)
	if err != nil {
		return uuid.Nil, err
	}
	skillDefs, err := r.ListSkillDefinitions(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	for _, sd := range skillDefs {
		cSkills = append(cSkills, models.CharacterSkillTO{
			ID:          uuid.New(),
			CharacterID: newID,
			SkillID:     sd.ID,
		})
	}
	e := r.withTx(ctx, func(tx *sqlx.Tx) error {
		return replaceAll(ctx, tx, skillTable, newID, cSkills)
	})
	return newID, e
}

func (r *DBCharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*CharacterAggregate, error) {
	c := models.CharacterTO{}
	if err := r.db.GetContext(ctx, &c, `SELECT * FROM character WHERE id = ?`, id); err != nil {
		return nil, err
	}
	agg := &CharacterAggregate{Character: &c}
	if ab, err := getOne[models.AbilitiesTO](ctx, r.db, "abilities", id); err != nil {
		return nil, err
	} else {
		agg.Abilities = ab
	}
	if st, err := getOne[models.SavingThrowsTO](ctx, r.db, "saving_throws", id); err != nil {
		return nil, err
	} else {
		agg.SavingThrows = st
	}
	if w, err := getOne[models.WalletTO](ctx, r.db, "wallet", id); err != nil {
		return nil, err
	} else {
		agg.Wallet = w
	}
	if items, err := selectAll(ctx, r.db, itemTable, id); err != nil {
		return nil, err
	} else {
		agg.Items = items
	}
	if spells, err := selectAll(ctx, r.db, spellTable, id); err != nil {
		return nil, err
	} else {
		agg.Spells = spells
	}
	if atks, err := selectAll(ctx, r.db, attackTable, id); err != nil {
		return nil, err
	} else {
		agg.Attacks = atks
	}
	if feats, err := selectAll(ctx, r.db, featureTable, id); err != nil {
		return nil, err
	} else {
		agg.Features = feats
	}
	if notes, err := selectAll(ctx, r.db, noteTable, id); err != nil {
		return nil, err
	} else {
		agg.Notes = notes
	}
	if skills, err := r.ListSkillDetailsByCharacter(ctx, id); err != nil {
		return nil, err
	} else {
		agg.Skills = skills
	}
	agg.shadow = agg.Clone()
	return agg, nil
}

func (r *DBCharacterRepository) ListSummary(ctx context.Context) ([]models.CharacterSummary, error) {
	var list []models.CharacterSummary
	if err := r.db.SelectContext(ctx, &list, `SELECT id, name FROM character ORDER BY name ASC`); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DBCharacterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	childTables := []string{
		"wallet", "abilities", "saving_throws",
		"item", "spell", "attacks", "character_skill", "features", "notes",
	}
	err := r.withTx(ctx, func(tx *sqlx.Tx) error {
		for _, name := range childTables {
			if err := deleteByCharacter(ctx, tx, name, id); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM character WHERE id=?`, id); err != nil {
			return err
		}
		return nil
	})
	return err
}

// Update persists the aggregate. When a shadow snapshot is present
// only sections that differ from the shadow are rewritten.
func (r *DBCharacterRepository) Update(ctx context.Context, agg *CharacterAggregate) error {
	if agg == nil || agg.Character == nil {
		return errors.New("Update: nil aggregate or character")
	}
	id := agg.Character.ID
	skills := util.Map(agg.Skills, func(s models.CharacterSkillDetailTO) models.CharacterSkillTO { return s.ToCharacterSkillTO() })

	err := r.withTx(ctx, func(tx *sqlx.Tx) error {
		c := agg.Character
		ensureSpellSlots(c)
		query := `
			UPDATE character SET
				name=?, class_levels=?, race=?, alignment=?,
				proficiency_bonus=?, armor_class=?, initiative=?, speed=?,
				max_hit_points=?, curr_hit_points=?, temp_hit_points=?,
				hit_dice=?, used_hit_dice=?, death_save_successes=?, death_save_failures=?,
				actions=?, bonus_actions=?, spell_slots=?, spell_slots_used=?,
				spellcasting_ability=?, spell_save_dc=?, spell_attack_bonus=?,
				age=?, height=?, weight=?, eyes=?, skin=?, hair=?, appearance=?,
				backstory=?, personality=?, updated_at = current_timestamp
			WHERE id=?
		`
		if _, err := tx.ExecContext(ctx, query,
			c.Name, c.ClassLevels, c.Race, c.Alignment,
			c.ProficiencyBonus, c.ArmorClass, c.Initiative, c.Speed,
			c.MaxHitPoints, c.CurrHitPoints, c.TempHitPoints,
			c.HitDice, c.UsedHitDice, c.DeathSaveSuccesses, c.DeathSaveFailures,
			c.Actions, c.BonusActions, c.SpellSlots, c.SpellSlotsUsed,
			c.SpellcastingAbility, c.SpellSaveDC, c.SpellAttackBonus,
			c.Age, c.Height, c.Weight, c.Eyes, c.Skin, c.Hair, c.Appearance, c.Backstory,
			c.Personality,
			c.ID,
		); err != nil {
			return err
		}

		shadow := agg.shadow
		var shadowSkills []models.CharacterSkillTO
		if shadow != nil {
			shadowSkills = util.Map(shadow.Skills, func(s models.CharacterSkillDetailTO) models.CharacterSkillTO { return s.ToCharacterSkillTO() })
		}

		// Owned (1:1) sections.
		if shadow == nil || !reflect.DeepEqual(agg.Abilities, shadow.Abilities) {
			if err := upsertOne(ctx, tx, abilitiesTable, id, agg.Abilities); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.SavingThrows, shadow.SavingThrows) {
			if err := upsertOne(ctx, tx, savingThrowsTable, id, agg.SavingThrows); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.Wallet, shadow.Wallet) {
			if err := upsertOne(ctx, tx, walletTable, id, agg.Wallet); err != nil {
				return err
			}
		}

		// Child (1:N) sections.
		if shadow == nil || !reflect.DeepEqual(agg.Items, shadow.Items) {
			if err := replaceAll(ctx, tx, itemTable, id, agg.Items); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.Spells, shadow.Spells) {
			if err := replaceAll(ctx, tx, spellTable, id, agg.Spells); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.Attacks, shadow.Attacks) {
			if err := replaceAll(ctx, tx, attackTable, id, agg.Attacks); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.Features, shadow.Features) {
			if err := replaceAll(ctx, tx, featureTable, id, agg.Features); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(agg.Notes, shadow.Notes) {
			if err := replaceAll(ctx, tx, noteTable, id, agg.Notes); err != nil {
				return err
			}
		}
		if shadow == nil || !reflect.DeepEqual(skills, shadowSkills) {
			if err := replaceAll(ctx, tx, skillTable, id, skills); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		agg.shadow = agg.Clone()
	}
	return err
}

// Helpers

func (r *DBCharacterRepository) withTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func ensureSpellSlots(c *models.CharacterTO) {
	if c.SpellSlots == nil || len(c.SpellSlots) != 10 {
		c.SpellSlots = make(models.IntList, 10)
	}
	if c.SpellSlotsUsed == nil || len(c.SpellSlotsUsed) != 10 {
		c.SpellSlotsUsed = make(models.IntList, 10)
	}
}

func (r *DBCharacterRepository) ListSkillDefinitions(ctx context.Context) ([]models.SkillDefinitionTO, error) {
	var defs []models.SkillDefinitionTO
	if err := r.db.SelectContext(ctx, &defs, `SELECT * FROM skill_definition ORDER BY id ASC`); err != nil {
		return nil, err
	}
	return defs, nil
}

func (r *DBCharacterRepository) ListSkillDetailsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillDetailTO, error) {
	var rows []models.CharacterSkillDetailTO
	query := `
		SELECT
			cs.id,
			cs.character_id,
			cs.skill_id,
			cs.proficiency,
			cs.custom_modifier,
			sd.name AS skill_name,
			sd.ability AS skill_ability,
			cs.created_at,
			cs.updated_at
		FROM character_skill cs
		JOIN skill_definition sd ON sd.id = cs.skill_id
		WHERE cs.character_id = ?
		ORDER BY cs.skill_id ASC
	`
	if err := r.db.SelectContext(ctx, &rows, query, characterID); err != nil {
		return nil, err
	}
	return rows, nil
}
