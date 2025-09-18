-- Ensure target names are unique within a mission
CREATE UNIQUE INDEX IF NOT EXISTS ux_targets_mission_name
ON targets(mission_id, name);

