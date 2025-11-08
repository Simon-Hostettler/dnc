-- +duckUp

-- duckDB does not support adding columns with constraints yet???
-- https://github.com/duckdb/duckdb/issues/3248
-- also does not allow more than one statement per alter

ALTER TABLE character
DROP COLUMN background;
ALTER TABLE character
ADD COLUMN age INTEGER DEFAULT 0;
ALTER TABLE character
ADD COLUMN height TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN weight TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN eyes TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN skin TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN hair TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN appearance TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN backstory TEXT DEFAULT '';
ALTER TABLE character
ADD COLUMN personality TEXT DEFAULT '';

CREATE TABLE IF NOT EXISTS features (
    id UUID PRIMARY KEY DEFAULT uuid(),
    character_id UUID NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
);

-- +duckDown
DROP TABLE IF EXISTS features;


ALTER TABLE character
ADD COLUMN background TEXT DEFAULT '';
ALTER TABLE character
DROP age;
ALTER TABLE character
DROP height;
ALTER TABLE character
DROP weight;
ALTER TABLE character
DROP eyes;
ALTER TABLE character
DROP skin;
ALTER TABLE character
DROP hair;
ALTER TABLE character
DROP appearance;
ALTER TABLE character
DROP backstory;
ALTER TABLE character
DROP personality;