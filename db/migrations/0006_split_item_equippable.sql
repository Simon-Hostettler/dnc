-- +duckUp

-- Recreate table since DuckDB doesn't support mixing adding/dropping/renaming columns and modifying data in the same transaction.
CREATE TABLE item_new (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL,
    is_equippable INTEGER NOT NULL DEFAULT 0 CHECK (is_equippable BETWEEN 0 AND 1),
    equipped INTEGER NOT NULL DEFAULT 0 CHECK (equipped BETWEEN 0 AND 1),
    attunement_slots INTEGER NOT NULL DEFAULT 0,
    quantity INTEGER NOT NULL DEFAULT 1,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

INSERT INTO item_new (id, character_id, name, is_equippable, equipped, attunement_slots, quantity, description, created_at, updated_at)
SELECT
    id,
    character_id,
    name,
    CASE WHEN equipped IN (1, 2) THEN 1 ELSE 0 END,
    CASE WHEN equipped = 2 THEN 1 ELSE 0 END,
    attunement_slots,
    quantity,
    description,
    created_at,
    updated_at
FROM item;

DROP TABLE item;
ALTER TABLE item_new RENAME TO item;

-- +duckDown

CREATE TABLE item_old (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL,
    equipped INTEGER NOT NULL DEFAULT 0 CHECK (equipped BETWEEN 0 AND 2),
    attunement_slots INTEGER NOT NULL DEFAULT 0,
    quantity INTEGER NOT NULL DEFAULT 1,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

INSERT INTO item_old (id, character_id, name, equipped, attunement_slots, quantity, description, created_at, updated_at)
SELECT
    id,
    character_id,
    name,
    CASE
        WHEN is_equippable = 0 THEN 0
        WHEN equipped     = 1 THEN 2
        ELSE 1
    END,
    attunement_slots,
    quantity,
    description,
    created_at,
    updated_at
FROM item;

DROP TABLE item;
ALTER TABLE item_old RENAME TO item;
