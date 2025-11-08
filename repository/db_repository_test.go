package repository

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"hostettler.dev/dnc/db"
	"hostettler.dev/dnc/models"
)

func TestEmptyOperations(t *testing.T) {
	dbPath := db.TestDBPath()
	handle, err := db.TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := db.MigrateUp(handle); err != nil {
		t.Fatalf("Migration to current version failed: %s", err.Error())
	}
	repo := NewDBCharacterRepository(handle)
	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create a new character: %s", err.Error())
	}
	c, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not retrieve the character: %s", err.Error())
	}
	err = repo.Update(ctx, c)
	if err != nil {
		t.Errorf("Could not update the character without changes: %s", err.Error())
	}
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Errorf("Could not delete the character: %s", err.Error())
	}
	cancel() // just to be sure
	if err := db.DestroyTestDB(handle, dbPath); err != nil {
		t.Fatalf("Could not destroy test DB: %s", err.Error())
	}
}

func TestAllValuesPersist(t *testing.T) {
	dbPath := db.TestDBPath()
	handle, err := db.TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := db.MigrateUp(handle); err != nil {
		t.Fatalf("Migration to current version failed: %s", err.Error())
	}
	repo := NewDBCharacterRepository(handle)
	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create a new character: %s", err.Error())
	}
	testChar := testCharacterAgg(id)
	err = repo.Update(ctx, &testChar)
	if err != nil {
		t.Fatalf("Could not update the character: %s", err.Error())
	}
	loaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not fetch the character: %s", err.Error())
	}
	if diff := cmp.Diff(testChar, *loaded, diffIgnoringTimestampsOption()); diff != "" {
		t.Errorf("Mismatch between stored and loaded values in character:\n%s", diff)
	}
	cancel() // just to be sure
	if err := db.DestroyTestDB(handle, dbPath); err != nil {
		t.Fatalf("Could not destroy test DB: %s", err.Error())
	}
}

func testCharacterAgg(id uuid.UUID) CharacterAggregate {
	return CharacterAggregate{
		Character: &models.CharacterTO{
			ID:                  id,
			Name:                "Bobby",
			ClassLevels:         "Wizard 10",
			Race:                "Gnome",
			Alignment:           "Chaotic Evil",
			ProficiencyBonus:    4,
			ArmorClass:          17,
			Initiative:          4,
			Speed:               30,
			MaxHitPoints:        100,
			CurrHitPoints:       50,
			TempHitPoints:       10,
			HitDice:             "10d6",
			UsedHitDice:         "5",
			DeathSaveSuccesses:  2,
			DeathSaveFailures:   2,
			Actions:             "Kick",
			BonusActions:        "Jump",
			SpellSlots:          []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			SpellSlotsUsed:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			SpellcastingAbility: "Cha",
			SpellSaveDC:         20,
			SpellAttackBonus:    10,
			Age:                 42,
			Height:              "4'11",
			Weight:              "200lbs",
			Eyes:                "Green",
			Skin:                "Pale",
			Hair:                "Brown",
			Appearance:          "Disheveled",
			Backstory:           "Noone",
			Personality:         "Crazy",
		},
		Abilities: &models.AbilitiesTO{
			CharacterID:  id,
			Strength:     10,
			Dexterity:    10,
			Constitution: 10,
			Intelligence: 10,
			Wisdom:       10,
			Charisma:     10,
		},
		SavingThrows: &models.SavingThrowsTO{
			CharacterID:             id,
			StrengthProficiency:     0,
			DexterityProficiency:    1,
			ConstitutionProficiency: 2,
			IntelligenceProficiency: 2,
			WisdomProficiency:       1,
			CharismaProficiency:     0,
		},
		Wallet: &models.WalletTO{
			CharacterID: id,
			Copper:      10,
			Silver:      20,
			Electrum:    30,
			Gold:        40,
			Platinum:    50,
		},
		Spells: []models.SpellTO{{
			ID:          uuid.New(),
			CharacterID: id,
			Name:        "Abracadabra",
			Level:       9,
			Prepared:    1,
			Damage:      "42d8",
			CastingTime: "Action",
			Range:       "600ft",
			Duration:    "Instant",
			Components:  "V/S/M",
			Description: "Boom",
		}},
		Items: []models.ItemTO{{
			ID:              uuid.New(),
			CharacterID:     id,
			Name:            "Stick",
			Equipped:        2,
			AttunementSlots: 3,
			Quantity:        1,
			Description:     "Stick",
		}},
		Attacks: []models.AttackTO{{
			ID:          uuid.New(),
			CharacterID: id,
			Name:        "Hit w/ Stick",
			Bonus:       15,
			Damage:      "10d6",
			DamageType:  "Necrotic",
		}},
		Skills: []models.CharacterSkillDetailTO{{
			ID:             uuid.New(),
			CharacterID:    id,
			SkillID:        1,
			Proficiency:    1,
			CustomModifier: 3,
			SkillName:      "Athletics",
			SkillAbility:   "Strength",
		}},
		Features: []models.FeatureTO{{
			ID:          uuid.New(),
			CharacterID: id,
			Name:        "Winged",
			Description: "Can fly",
		}},
	}
}

func diffIgnoringTimestampsOption() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		sf, ok := p.Last().(cmp.StructField)
		if !ok {
			return false
		}
		name := sf.Name()
		return name == "CreatedAt" || name == "UpdatedAt"
	}, cmp.Ignore())
}
