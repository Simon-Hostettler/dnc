package repository

import (
	"context"
	"database/sql"
	"errors"

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
                name, class_levels, background, alignment,
                proficiency_bonus, armor_class, initiative, speed,
                max_hit_points, curr_hit_points, temp_hit_points,
                hit_dice, used_hit_dice, death_save_successes, death_save_failures,
				actions, bonus_actions, spell_slots, spell_slots_used,
                spellcasting_ability, spell_save_dc, spell_attack_bonus
            ) VALUES (
                ?,?,?,?,?,?,?,?,?,?,?,
                ?,?,?,?,?,?,?,?,?,?,?
			) RETURNING id`
		row := tx.QueryRowxContext(ctx, query,
			c.Name, c.ClassLevels, c.Background, c.Alignment,
			c.ProficiencyBonus, c.ArmorClass, c.Initiative, c.Speed,
			c.MaxHitPoints, c.CurrHitPoints, c.TempHitPoints,
			c.HitDice, c.UsedHitDice, c.DeathSaveSuccesses, c.DeathSaveFailures,
			c.Actions, c.BonusActions, c.SpellSlots, c.SpellSlotsUsed,
			c.SpellcastingAbility, c.SpellSaveDC, c.SpellAttackBonus,
		)
		if err := row.Scan(&newID); err != nil {
			return err
		}

		if agg.Abilities != nil {
			if err := upsertAbilities(ctx, tx, newID, agg.Abilities); err != nil {
				return err
			}
		}
		if agg.SavingThrows != nil {
			if err := upsertSavingThrows(ctx, tx, newID, agg.SavingThrows); err != nil {
				return err
			}
		}
		if agg.Wallet != nil {
			if err := upsertWallet(ctx, tx, newID, agg.Wallet); err != nil {
				return err
			}
		}

		if len(agg.Items) > 0 {
			if err := replaceItems(ctx, tx, newID, agg.Items); err != nil {
				return err
			}
		}
		if len(agg.Spells) > 0 {
			if err := replaceSpells(ctx, tx, newID, agg.Spells); err != nil {
				return err
			}
		}
		if len(agg.Attacks) > 0 {
			if err := replaceAttacks(ctx, tx, newID, agg.Attacks); err != nil {
				return err
			}
		}
		skills := util.Map(agg.Skills, func(s models.CharacterSkillDetailTO) models.CharacterSkillTO { return s.ToCharacterSkillTO() })
		if len(skills) > 0 {
			if err := replaceSkills(ctx, tx, newID, skills); err != nil {
				return err
			}
		}
		agg.Character.ID = newID
		return nil
	})
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
		if err := replaceSkills(ctx, tx, newID, cSkills); err != nil {
			return err
		}
		return nil
	})
	return newID, e
}

func (r *DBCharacterRepository) GetByID(ctx context.Context, id uuid.UUID) (*CharacterAggregate, error) {
	c := models.CharacterTO{}
	if err := r.db.GetContext(ctx, &c, `SELECT * FROM character WHERE id = ?`, id); err != nil {
		return nil, err
	}
	agg := &CharacterAggregate{Character: &c}
	if ab, err := getAbilities(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.Abilities = ab
	}
	if st, err := getSavingThrows(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.SavingThrows = st
	}
	if w, err := getWallet(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.Wallet = w
	}
	if items, err := listItems(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.Items = items
	}
	if spells, err := listSpells(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.Spells = spells
	}
	if atks, err := listAttacks(ctx, r.db, id); err != nil {
		return nil, err
	} else {
		agg.Attacks = atks
	}
	if skills, err := r.ListSkillDetailsByCharacter(ctx, id); err != nil {
		return nil, err
	} else {
		agg.Skills = skills
	}
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
	err := r.withTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM wallet WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM abilities WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM saving_throws WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM item WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM spell WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM attacks WHERE character_id=?`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM character_skill WHERE character_id=?`, id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM character WHERE id=?`, id); err != nil {
			return err
		}
		return nil
	})
}

// Update persists the aggregate: character row plus all 1:1 and 1:N dependents.
func (r *DBCharacterRepository) Update(ctx context.Context, agg *CharacterAggregate) error {
	if agg == nil || agg.Character == nil {
		return errors.New("Update: nil aggregate or character")
	}
	id := agg.Character.ID
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		// Update character row
		c := agg.Character
		ensureSpellSlots(c)
		query := `
			UPDATE character SET
				name=?, class_levels=?, background=?, alignment=?,
				proficiency_bonus=?, armor_class=?, initiative=?, speed=?,
				max_hit_points=?, curr_hit_points=?, temp_hit_points=?,
				hit_dice=?, used_hit_dice=?, death_save_successes=?, death_save_failures=?,
				actions=?, bonus_actions=?, spell_slots=?, spell_slots_used=?,
				spellcasting_ability=?, spell_save_dc=?, spell_attack_bonus=?,
				updated_at = current_timestamp
			WHERE id=?
		`
		if _, err := tx.ExecContext(ctx, query,
			c.Name, c.ClassLevels, c.Background, c.Alignment,
			c.ProficiencyBonus, c.ArmorClass, c.Initiative, c.Speed,
			c.MaxHitPoints, c.CurrHitPoints, c.TempHitPoints,
			c.HitDice, c.UsedHitDice, c.DeathSaveSuccesses, c.DeathSaveFailures,
			c.Actions, c.BonusActions, c.SpellSlots, c.SpellSlotsUsed,
			c.SpellcastingAbility, c.SpellSaveDC, c.SpellAttackBonus,
			c.ID,
		); err != nil {
			return err
		}

		// 1:1 tables
		if err := upsertAbilities(ctx, tx, id, agg.Abilities); err != nil {
			return err
		}
		if err := upsertSavingThrows(ctx, tx, id, agg.SavingThrows); err != nil {
			return err
		}
		if err := upsertWallet(ctx, tx, id, agg.Wallet); err != nil {
			return err
		}

		// 1:N tables (replace strategy)
		if err := replaceItems(ctx, tx, id, agg.Items); err != nil {
			return err
		}
		if err := replaceSpells(ctx, tx, id, agg.Spells); err != nil {
			return err
		}
		if err := replaceAttacks(ctx, tx, id, agg.Attacks); err != nil {
			return err
		}
		skills := util.Map(agg.Skills, func(s models.CharacterSkillDetailTO) models.CharacterSkillTO { return s.ToCharacterSkillTO() })
		if err := replaceSkills(ctx, tx, id, skills); err != nil {
			return err
		}
		return nil
	})
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

func upsertAbilities(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, ab *models.AbilitiesTO) error {
	if ab == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM abilities WHERE character_id=?`, characterID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
        INSERT INTO abilities (
            character_id, strength, dexterity, constitution, intelligence, wisdom, charisma
        ) VALUES (?,?,?,?,?,?,?)
    `, characterID, ab.Strength, ab.Dexterity, ab.Constitution, ab.Intelligence, ab.Wisdom, ab.Charisma)
	return err
}

func upsertSavingThrows(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, st *models.SavingThrowsTO) error {
	if st == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM saving_throws WHERE character_id=?`, characterID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
        INSERT INTO saving_throws (
            character_id, strength_proficiency, dexterity_proficiency, constitution_proficiency, intelligence_proficiency, wisdom_proficiency, charisma_proficiency
        ) VALUES (?,?,?,?,?,?,?)
    `, characterID, st.StrengthProficiency, st.DexterityProficiency, st.ConstitutionProficiency, st.IntelligenceProficiency, st.WisdomProficiency, st.CharismaProficiency)
	return err
}

func upsertWallet(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, w *models.WalletTO) error {
	if w == nil {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM wallet WHERE character_id=?`, characterID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
        INSERT INTO wallet (
            character_id, copper, silver, electrum, gold, platinum
        ) VALUES (?,?,?,?,?,?)
    `, characterID, w.Copper, w.Silver, w.Electrum, w.Gold, w.Platinum)
	return err
}

func replaceItems(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, items []models.ItemTO) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM item WHERE character_id=?`, characterID); err != nil {
		return err
	}
	for _, it := range items {
		if it.ID == uuid.Nil {
			it.ID = uuid.New()
		}
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO item (id, character_id, name, equipped, attunement_slots, quantity)
            VALUES (?,?,?,?,?,?)
        `, it.ID, characterID, it.Name, it.Equipped, it.AttunementSlots, it.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func replaceSpells(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, spells []models.SpellTO) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM spell WHERE character_id=?`, characterID); err != nil {
		return err
	}
	for _, sp := range spells {
		if sp.ID == uuid.Nil {
			sp.ID = uuid.New()
		}
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO spell (id, character_id, name, level, prepared, damage, casting_time, range, duration, components, description)
            VALUES (?,?,?,?,?,?,?,?,?,?,?)
        `, sp.ID, characterID, sp.Name, sp.Level, sp.Prepared, sp.Damage, sp.CastingTime, sp.Range, sp.Duration, sp.Components, sp.Description); err != nil {
			return err
		}
	}
	return nil
}

func replaceAttacks(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, atks []models.AttackTO) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM attacks WHERE character_id=?`, characterID); err != nil {
		return err
	}
	for _, a := range atks {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO attacks (id, character_id, name, bonus, damage, damage_type)
            VALUES (?,?,?,?,?,?)
        `, a.ID, characterID, a.Name, a.Bonus, a.Damage, a.DamageType); err != nil {
			return err
		}
	}
	return nil
}

func replaceSkills(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, skills []models.CharacterSkillTO) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM character_skill WHERE character_id=?`, characterID); err != nil {
		return err
	}
	for _, s := range skills {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO character_skill (id, character_id, skill_id, proficiency, custom_modifier)
            VALUES (?,?,?,?,?)
        `, s.ID, characterID, s.SkillID, s.Proficiency, s.CustomModifier); err != nil {
			return err
		}
	}
	return nil
}

func getAbilities(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) (*models.AbilitiesTO, error) {
	var ab models.AbilitiesTO
	if err := sqlx.GetContext(ctx, db, &ab, `SELECT * FROM abilities WHERE character_id=?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ab, nil
}

func getSavingThrows(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) (*models.SavingThrowsTO, error) {
	var st models.SavingThrowsTO
	if err := sqlx.GetContext(ctx, db, &st, `SELECT * FROM saving_throws WHERE character_id=?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &st, nil
}

func getWallet(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) (*models.WalletTO, error) {
	var w models.WalletTO
	if err := sqlx.GetContext(ctx, db, &w, `SELECT * FROM wallet WHERE character_id=?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func listItems(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) ([]models.ItemTO, error) {
	var items []models.ItemTO
	if err := sqlx.SelectContext(ctx, db, &items, `SELECT * FROM item WHERE character_id=? ORDER BY created_at ASC`, id); err != nil {
		return nil, err
	}
	return items, nil
}

func listSpells(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) ([]models.SpellTO, error) {
	var spells []models.SpellTO
	if err := sqlx.SelectContext(ctx, db, &spells, `SELECT * FROM spell WHERE character_id=? ORDER BY level ASC, name ASC`, id); err != nil {
		return nil, err
	}
	return spells, nil
}

func listAttacks(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) ([]models.AttackTO, error) {
	var atks []models.AttackTO
	if err := sqlx.SelectContext(ctx, db, &atks, `SELECT * FROM attacks WHERE character_id=? ORDER BY created_at ASC`, id); err != nil {
		return nil, err
	}
	return atks, nil
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
