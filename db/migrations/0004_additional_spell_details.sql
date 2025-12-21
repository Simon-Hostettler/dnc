-- +duckUp

ALTER TABLE spell
ADD COLUMN school TEXT DEFAULT '';
ALTER TABLE spell
ADD COLUMN spell_source INTEGER DEFAULT 0;

-- +duckDown

ALTER TABLE spell
DROP school;
ALTER TABLE spell
DROP spell_source;