-- +duckUp

ALTER TABLE spell
ADD COLUMN concentration INTEGER DEFAULT 0;
ALTER TABLE spell
ADD COLUMN ritual INTEGER DEFAULT 0;

-- +duckDown

ALTER TABLE spell
DROP concentration;
ALTER TABLE spell
DROP ritual;