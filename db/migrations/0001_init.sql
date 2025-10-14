-- +duckUp
CREATE TABLE IF NOT EXISTS character (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    class_levels TEXT NOT NULL DEFAULT '',
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
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS item (
    id TEXT PRIMARY KEY,
    character_id TEXT NOT NULL,
    name TEXT NOT NULL,
    equipped INTEGER NOT NULL DEFAULT 0 CHECK (
        equipped BETWEEN 0
        AND 2
    ),
    attunement_slots INTEGER NOT NULL DEFAULT 0,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    FOREIGN KEY(character_id) REFERENCES character(id) ON
    DELETE
        CASCADE
);

CREATE TABLE IF NOT EXISTS wallet (
    character_id TEXT PRIMARY KEY,
    copper INTEGER NOT NULL DEFAULT 0,
    silver INTEGER NOT NULL DEFAULT 0,
    electrum INTEGER NOT NULL DEFAULT 0,
    gold INTEGER NOT NULL DEFAULT 0,
    platinum INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    FOREIGN KEY(character_id) REFERENCES character(id) ON
    DELETE
        CASCADE
);

CREATE TABLE IF NOT EXISTS spell (
    id TEXT PRIMARY KEY,
    character_id TEXT NOT NULL,
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
    FOREIGN KEY(character_id) REFERENCES character(id) ON
    DELETE
        CASCADE
);

-- +duckDown
DROP TABLE IF EXISTS spell;

DROP TABLE IF EXISTS wallet;

DROP TABLE IF EXISTS item;

DROP TABLE IF EXISTS character;