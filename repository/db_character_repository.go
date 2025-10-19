package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/util"
)

// DBCharacterRepository is a CharacterRepository backed by sqlx and DuckDB.
type DBCharacterRepository struct {
	db *sqlx.DB
}

func NewDBCharacterRepository(db *sqlx.DB) *DBCharacterRepository {
	return &DBCharacterRepository{db: db}
}

var (
	colMapsOnce sync.Once
	colMaps     map[string]map[string]string
)

// returns loose key->column mapping for the given table.
// Keys supported (case-insensitive):
// - db tag (e.g. "name")
// - json tag (e.g. "class_levels")
// - snake_case of field name (e.g. "ClassLevels" -> "class_levels")
func columnMap(table string) map[string]string {
	colMapsOnce.Do(buildAllColumnMaps)
	return colMaps[table]
}

func buildAllColumnMaps() {
	colMaps = make(map[string]map[string]string)

	commonExclude := map[string]struct{}{
		"id":         {},
		"created_at": {},
		"updated_at": {},
	}

	register := func(table string, t reflect.Type, exclude map[string]struct{}) {
		if exclude == nil {
			exclude = commonExclude
		}
		colMaps[table] = buildColumnMap(t, exclude)
	}

	register("character", reflect.TypeOf(models.CharacterTO{}), commonExclude)
	register("abilities", reflect.TypeOf(models.AbilitiesTO{}), commonExclude)
	register("saving_throws", reflect.TypeOf(models.SavingThrowsTO{}), commonExclude)
	register("wallet", reflect.TypeOf(models.WalletTO{}), commonExclude)
	register("item", reflect.TypeOf(models.ItemTO{}), commonExclude)
	register("spell", reflect.TypeOf(models.SpellTO{}), commonExclude)
	register("attacks", reflect.TypeOf(models.AttackTO{}), commonExclude)
	register("character_skill", reflect.TypeOf(models.CharacterSkillTO{}), commonExclude)
	register("skill_definition", reflect.TypeOf(models.SkillDefinitionTO{}), commonExclude)
}

func buildColumnMap(t reflect.Type, exclude map[string]struct{}) map[string]string {
	m := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		dbTag := f.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		if _, skip := exclude[dbTag]; skip {
			continue
		}

		// db tag
		addKey(m, dbTag, dbTag)

		// json tag (first part before comma)
		if jt := f.Tag.Get("json"); jt != "" && jt != "-" {
			if idx := strings.IndexByte(jt, ','); idx >= 0 {
				jt = jt[:idx]
			}
			if jt != "" {
				addKey(m, jt, dbTag)
			}
		}

		// snake_case of field name
		addKey(m, snakeCase(f.Name), dbTag)
	}
	return m
}

func addKey(m map[string]string, key, col string) {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		return
	}
	m[k] = col
}

func snakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (r *DBCharacterRepository) updateTableFields(ctx context.Context, table string, identifier string, id uuid.UUID, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	mapping := columnMap(table)
	if mapping == nil {
		return fmt.Errorf("no column mapping for table %q", table)
	}

	setParts := make([]string, 0, len(fields)+1)
	args := make([]any, 0, len(fields)+1)
	for k, v := range fields {
		key := strings.ToLower(strings.TrimSpace(k))
		col, ok := mapping[key]
		if !ok {
			return fmt.Errorf("UpdateTableFields: field %q not allowed on table %q", k, table)
		}
		setParts = append(setParts, col+" = ?")
		args = append(args, v)
	}
	setParts = append(setParts, "updated_at = current_timestamp")

	query := "UPDATE " + table + " SET " + strings.Join(setParts, ", ") + " WHERE " + identifier + " = ?"
	args = append(args, id)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return nil
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
	var c models.CharacterTO
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
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
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
		if _, err := tx.ExecContext(ctx, `DELETE FROM character WHERE id=?`, id); err != nil {
			return err
		}
		return nil
	})
}

func (r *DBCharacterRepository) UpdateCharacter(ctx context.Context, c models.CharacterTO) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error {
		ensureSpellSlots(&c)
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
		return nil
	})
}

func (r *DBCharacterRepository) UpdateSpellSlotsMax(ctx context.Context, characterID uuid.UUID, level int, v int) error {
	if level < 0 || level >= 10 {
		return errors.New("UpdateSpellSlot: level out of range [0..9]")
	}
	var c models.CharacterTO
	if err := r.db.GetContext(ctx, &c, `SELECT id, spell_slots, spell_slots_used FROM character WHERE id=?`, characterID); err != nil {
		return err
	}
	ensureSpellSlots(&c)

	c.SpellSlots[level] = v
	query := `UPDATE character SET spell_slots=?, updated_at=current_timestamp WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, c.SpellSlots, characterID)
	return err
}

func (r *DBCharacterRepository) UpdateSpellSlotsUsed(ctx context.Context, characterID uuid.UUID, level int, v int) error {
	if level < 0 || level >= 10 {
		return errors.New("UpdateSpellSlot: level out of range [0..9]")
	}
	var c models.CharacterTO
	if err := r.db.GetContext(ctx, &c, `SELECT id, spell_slots, spell_slots_used FROM character WHERE id=?`, characterID); err != nil {
		return err
	}
	ensureSpellSlots(&c)

	c.SpellSlotsUsed[level] = v
	query := `UPDATE character SET spell_slots_used=?, updated_at=current_timestamp WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, c.SpellSlotsUsed, characterID)
	return err
}

func (r *DBCharacterRepository) UpdateCharacterFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	print(fields)
	return r.updateTableFields(ctx, "character", "id", id, fields)
}

func (r *DBCharacterRepository) ListSkillDefinitions(ctx context.Context) ([]models.SkillDefinitionTO, error) {
	var defs []models.SkillDefinitionTO
	if err := r.db.SelectContext(ctx, &defs, `SELECT * FROM skill_definition ORDER BY id ASC`); err != nil {
		return nil, err
	}
	return defs, nil
}

func (r *DBCharacterRepository) AddSpell(ctx context.Context, characterID uuid.UUID, sp models.SpellTO) (uuid.UUID, error) {
	if sp.ID != uuid.Nil {
		var id uuid.UUID
		row := r.db.QueryRowxContext(ctx, `
            INSERT INTO spell (id, character_id, name, level, prepared, damage, casting_time, range, duration, components, description)
            VALUES (?,?,?,?,?,?,?,?,?,?,?) RETURNING id
        `, sp.ID, characterID, sp.Name, sp.Level, sp.Prepared, sp.Damage, sp.CastingTime, sp.Range, sp.Duration, sp.Components, sp.Description)
		if err := row.Scan(&id); err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}
	var id uuid.UUID
	row := r.db.QueryRowxContext(ctx, `
        INSERT INTO spell (character_id, name, level, prepared, damage, casting_time, range, duration, components, description)
        VALUES (?,?,?,?,?,?,?,?,?,?) RETURNING id
    `, characterID, sp.Name, sp.Level, sp.Prepared, sp.Damage, sp.CastingTime, sp.Range, sp.Duration, sp.Components, sp.Description)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *DBCharacterRepository) UpdateSpell(ctx context.Context, sp models.SpellTO) error {
	if sp.ID == uuid.Nil {
		return errors.New("UpdateSpell: missing spell ID")
	}
	_, err := r.db.ExecContext(ctx, `
        UPDATE spell SET
            name=?, level=?, prepared=?, damage=?, casting_time=?, range=?, duration=?, components=?, description=?,
            updated_at=current_timestamp
        WHERE id=?
    `, sp.Name, sp.Level, sp.Prepared, sp.Damage, sp.CastingTime, sp.Range, sp.Duration, sp.Components, sp.Description, sp.ID)
	return err
}

func (r *DBCharacterRepository) DeleteSpell(ctx context.Context, spellID uuid.UUID) error {
	if spellID == uuid.Nil {
		return errors.New("DeleteSpell: missing spell ID")
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM spell WHERE id=?`, spellID)
	return err
}

func (r *DBCharacterRepository) GetSpell(ctx context.Context, spellID uuid.UUID) (*models.SpellTO, error) {
	if spellID == uuid.Nil {
		return nil, errors.New("GetSpell: missing spell ID")
	}
	var sp models.SpellTO
	if err := r.db.GetContext(ctx, &sp, `SELECT * FROM spell WHERE id=?`, spellID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

func (r *DBCharacterRepository) UpdateSpellFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "spell", "id", id, fields)
}

func (r *DBCharacterRepository) ListSpellsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.SpellTO, error) {
	return listSpells(ctx, r.db, characterID)
}

func (r *DBCharacterRepository) AddItem(ctx context.Context, characterID uuid.UUID, it models.ItemTO) (uuid.UUID, error) {
	if it.ID != uuid.Nil {
		var id uuid.UUID
		row := r.db.QueryRowxContext(ctx, `
			INSERT INTO item (id, character_id, name, equipped, attunement_slots, quantity)
			VALUES (?,?,?,?,?,?) RETURNING id
		`, it.ID, characterID, it.Name, it.Equipped, it.AttunementSlots, it.Quantity)
		if err := row.Scan(&id); err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}
	var id uuid.UUID
	row := r.db.QueryRowxContext(ctx, `
		INSERT INTO item (character_id, name, equipped, attunement_slots, quantity)
		VALUES (?,?,?,?,?) RETURNING id
	`, characterID, it.Name, it.Equipped, it.AttunementSlots, it.Quantity)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *DBCharacterRepository) UpdateItem(ctx context.Context, it models.ItemTO) error {
	if it.ID == uuid.Nil {
		return errors.New("UpdateItem: missing item ID")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE item SET
			name=?, equipped=?, attunement_slots=?, quantity=?,
			updated_at=current_timestamp
		WHERE id=?
	`, it.Name, it.Equipped, it.AttunementSlots, it.Quantity, it.ID)
	return err
}

func (r *DBCharacterRepository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	if itemID == uuid.Nil {
		return errors.New("DeleteItem: missing item ID")
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM item WHERE id=?`, itemID)
	return err
}

func (r *DBCharacterRepository) GetItem(ctx context.Context, itemID uuid.UUID) (*models.ItemTO, error) {
	if itemID == uuid.Nil {
		return nil, errors.New("GetItem: missing item ID")
	}
	var it models.ItemTO
	if err := r.db.GetContext(ctx, &it, `SELECT * FROM item WHERE id=?`, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &it, nil
}

func (r *DBCharacterRepository) ListItemsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.ItemTO, error) {
	return listItems(ctx, r.db, characterID)
}

func (r *DBCharacterRepository) UpdateWalletFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "wallet", "character_id", id, fields)
}

func (r *DBCharacterRepository) UpdateItemFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "item", "id", id, fields)
}

func (r *DBCharacterRepository) AddAttack(ctx context.Context, characterID uuid.UUID, a models.AttackTO) (uuid.UUID, error) {
	if a.ID != uuid.Nil {
		var id uuid.UUID
		row := r.db.QueryRowxContext(ctx, `
			INSERT INTO attacks (id, character_id, name, bonus, damage, damage_type)
			VALUES (?,?,?,?,?,?) RETURNING id
		`, a.ID, characterID, a.Name, a.Bonus, a.Damage, a.DamageType)
		if err := row.Scan(&id); err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}
	var id uuid.UUID
	row := r.db.QueryRowxContext(ctx, `
		INSERT INTO attacks (character_id, name, bonus, damage, damage_type)
		VALUES (?,?,?,?,?) RETURNING id
	`, characterID, a.Name, a.Bonus, a.Damage, a.DamageType)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *DBCharacterRepository) UpdateAttack(ctx context.Context, a models.AttackTO) error {
	if a.ID == uuid.Nil {
		return errors.New("UpdateAttack: missing attack ID")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE attacks SET
			name=?, bonus=?, damage=?, damage_type=?,
			updated_at=current_timestamp
		WHERE id=?
	`, a.Name, a.Bonus, a.Damage, a.DamageType, a.ID)
	return err
}

func (r *DBCharacterRepository) UpdateAttackFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "attacks", "id", id, fields)
}

func (r *DBCharacterRepository) DeleteAttack(ctx context.Context, attackID uuid.UUID) error {
	if attackID == uuid.Nil {
		return errors.New("DeleteAttack: missing attack ID")
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM attacks WHERE id=?`, attackID)
	return err
}

func (r *DBCharacterRepository) GetAttack(ctx context.Context, attackID uuid.UUID) (*models.AttackTO, error) {
	if attackID == uuid.Nil {
		return nil, errors.New("GetAttack: missing attack ID")
	}
	var a models.AttackTO
	if err := r.db.GetContext(ctx, &a, `SELECT * FROM attacks WHERE id=?`, attackID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *DBCharacterRepository) ListAttacksByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.AttackTO, error) {
	return listAttacks(ctx, r.db, characterID)
}

func (r *DBCharacterRepository) UpsertSkill(ctx context.Context, characterID uuid.UUID, skillID int, proficiency int, customModifier int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM character_skill WHERE character_id=? AND skill_id=?`, characterID, skillID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO character_skill (id, character_id, skill_id, proficiency, custom_modifier)
		VALUES (?,?,?,?,?)
	`, uuid.New(), characterID, skillID, proficiency, customModifier)
	return err
}

func (r *DBCharacterRepository) UpdateSkillFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "character_skill", "skill_id", id, fields)
}

func (r *DBCharacterRepository) DeleteSkill(ctx context.Context, characterID uuid.UUID, skillID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM character_skill WHERE character_id=? AND skill_id=?`, characterID, skillID)
	return err
}

func (r *DBCharacterRepository) ListSkillsByCharacter(ctx context.Context, characterID uuid.UUID) ([]models.CharacterSkillTO, error) {
	return listSkills(ctx, r.db, characterID)
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

func (r *DBCharacterRepository) GetAbilities(ctx context.Context, characterID uuid.UUID) (*models.AbilitiesTO, error) {
	return getAbilities(ctx, r.db, characterID)
}

func (r *DBCharacterRepository) UpsertAbilities(ctx context.Context, characterID uuid.UUID, ab models.AbilitiesTO) error {
	return r.withTx(ctx, func(tx *sqlx.Tx) error { return upsertAbilities(ctx, tx, characterID, &ab) })
}

func (r *DBCharacterRepository) UpdateAbilityFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "abilities", "character_id", id, fields)
}

func (r *DBCharacterRepository) GetWallet(ctx context.Context, characterID uuid.UUID) (*models.WalletTO, error) {
	return getWallet(ctx, r.db, characterID)
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

func listSkills(ctx context.Context, db sqlx.QueryerContext, id uuid.UUID) ([]models.CharacterSkillTO, error) {
	var skills []models.CharacterSkillTO
	if err := sqlx.SelectContext(ctx, db, &skills, `SELECT * FROM character_skill WHERE character_id=? ORDER BY skill_id ASC`, id); err != nil {
		return nil, err
	}
	return skills, nil
}

func (r *DBCharacterRepository) UpdateSavingThrowFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.updateTableFields(ctx, "saving_throws", "character_id", id, fields)
}
