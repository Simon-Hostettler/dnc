-- +duckUp

-- Never use foreign keys in duckDB, will make updates impossible,
-- as they are implemented as insert + delete
CREATE TABLE IF NOT EXISTS character (
    id UUID PRIMARY KEY DEFAULT uuid(),
    name TEXT NOT NULL,
    class_levels TEXT NOT NULL DEFAULT '',
    race TEXT NOT NULL DEFAULT '',
    background TEXT NOT NULL DEFAULT '',
    alignment TEXT NOT NULL DEFAULT '',
    proficiency_bonus INTEGER NOT NULL,
    armor_class INTEGER NOT NULL,
    initiative INTEGER NOT NULL,
    speed INTEGER NOT NULL,
    max_hit_points INTEGER NOT NULL,
    curr_hit_points INTEGER NOT NULL,
    temp_hit_points INTEGER NOT NULL DEFAULT 0,
    hit_dice TEXT NOT NULL DEFAULT '',
    used_hit_dice TEXT NOT NULL DEFAULT '',
    death_save_successes INTEGER NOT NULL DEFAULT 0,
    death_save_failures INTEGER NOT NULL DEFAULT 0,
    actions TEXT NOT NULL DEFAULT '',
    bonus_actions TEXT NOT NULL DEFAULT '',
    spell_slots INTEGER [ 10 ] NOT NULL,
    spell_slots_used INTEGER [ 10 ] NOT NULL,
    spellcasting_ability TEXT NOT NULL DEFAULT '',
    spell_save_dc INT NOT NULL DEFAULT 0,
    spell_attack_bonus INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS item (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL,
    equipped INTEGER NOT NULL DEFAULT 0 CHECK (
        equipped BETWEEN 0
        AND 2
    ),
    attunement_slots INTEGER NOT NULL DEFAULT 0,
    quantity INTEGER NOT NULL DEFAULT 1,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS wallet (
    character_id UUID PRIMARY KEY,
    copper INTEGER NOT NULL DEFAULT 0,
    silver INTEGER NOT NULL DEFAULT 0,
    electrum INTEGER NOT NULL DEFAULT 0,
    gold INTEGER NOT NULL DEFAULT 0,
    platinum INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS spell (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL,
    level INTEGER NOT NULL,
    prepared INTEGER NOT NULL DEFAULT 0 CHECK (prepared IN (0, 1)),
    damage TEXT NOT NULL DEFAULT '',
    casting_time TEXT NOT NULL DEFAULT '',
    range TEXT NOT NULL DEFAULT '',
    duration TEXT NOT NULL DEFAULT '',
    components TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS attacks (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL,
    bonus INTEGER NOT NULL DEFAULT 0,
    damage TEXT NOT NULL,
    damage_type TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS abilities (
    character_id UUID PRIMARY KEY,
    strength INTEGER NOT NULL,
    dexterity INTEGER NOT NULL,
    constitution INTEGER NOT NULL,
    intelligence INTEGER NOT NULL,
    wisdom INTEGER NOT NULL,
    charisma INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS saving_throws (
    character_id UUID PRIMARY KEY,
    strength_proficiency INTEGER NOT NULL,
    dexterity_proficiency INTEGER NOT NULL,
    constitution_proficiency INTEGER NOT NULL,
    intelligence_proficiency INTEGER NOT NULL,
    wisdom_proficiency INTEGER NOT NULL,
    charisma_proficiency INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

CREATE TABLE IF NOT EXISTS skill_definition (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    ability TEXT NOT NULL CHECK (
        ability IN (
            'Strength',
            'Dexterity',
            'Constitution',
            'Intelligence',
            'Wisdom',
            'Charisma'
        )
    )
);

CREATE TABLE IF NOT EXISTS character_skill (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    skill_id INTEGER NOT NULL,
    proficiency INTEGER NOT NULL DEFAULT 0 CHECK (
        proficiency BETWEEN 0
        AND 2
    ),
    custom_modifier INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    UNIQUE(character_id, skill_id)
);

INSERT INTO
    skill_definition (id, name, ability)
VALUES
    (1, 'Athletics', 'Strength'),
    (2, 'Acrobatics', 'Dexterity'),
    (3, 'Sleight of Hand', 'Dexterity'),
    (4, 'Stealth', 'Dexterity'),
    (5, 'Arcana', 'Intelligence'),
    (6, 'History', 'Intelligence'),
    (7, 'Investigation', 'Intelligence'),
    (8, 'Nature', 'Intelligence'),
    (9, 'Religion', 'Intelligence'),
    (10, 'Animal Handling', 'Wisdom'),
    (11, 'Insight', 'Wisdom'),
    (12, 'Medicine', 'Wisdom'),
    (13, 'Perception', 'Wisdom'),
    (14, 'Survival', 'Wisdom'),
    (15, 'Deception', 'Charisma'),
    (16, 'Intimidation', 'Charisma'),
    (17, 'Performance', 'Charisma'),
    (18, 'Persuasion', 'Charisma');

-- +duckDown
DROP TABLE IF EXISTS spell;

DROP TABLE IF EXISTS wallet;

DROP TABLE IF EXISTS item;

DROP TABLE IF EXISTS attacks;

DROP TABLE IF EXISTS abilities;

DROP TABLE IF EXISTS character_skill;

DROP TABLE IF EXISTS skill_definition;

DROP TABLE IF EXISTS character;