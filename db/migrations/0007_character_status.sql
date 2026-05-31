-- +duckUp

ALTER TABLE character ADD COLUMN exhaustion INTEGER DEFAULT 0;
ALTER TABLE character ADD COLUMN concentration INTEGER DEFAULT 0;
ALTER TABLE character ADD COLUMN inspiration INTEGER DEFAULT 0;
ALTER TABLE character ADD COLUMN condition TEXT DEFAULT '';

-- +duckDown

ALTER TABLE character DROP exhaustion;
ALTER TABLE character DROP concentration;
ALTER TABLE character DROP inspiration;
ALTER TABLE character DROP condition;
