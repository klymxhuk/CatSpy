-- A cat may have at most one ACTIVE (not completed) mission
CREATE UNIQUE INDEX IF NOT EXISTS ux_active_mission_per_cat
ON missions(assigned_cat_id)
WHERE assigned_cat_id IS NOT NULL AND completed = false;


-- Targets per mission: enforce at most 3 via trigger (min 1 handled in handler)
CREATE OR REPLACE FUNCTION ensure_max_3_targets()
RETURNS trigger AS $$
BEGIN
IF (SELECT COUNT(*) FROM targets WHERE mission_id = NEW.mission_id) >= 3 THEN
RAISE EXCEPTION 'mission already has 3 targets';
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DO $$ BEGIN
IF NOT EXISTS (
SELECT 1 FROM pg_trigger WHERE tgname = 'tg_max_3_targets') THEN
CREATE TRIGGER tg_max_3_targets
BEFORE INSERT ON targets
FOR EACH ROW EXECUTE FUNCTION ensure_max_3_targets();
END IF;
END $$;
