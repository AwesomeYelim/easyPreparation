ALTER TABLE licenses ADD COLUMN last_verified TEXT;
ALTER TABLE licenses ADD COLUMN signature TEXT DEFAULT '';
