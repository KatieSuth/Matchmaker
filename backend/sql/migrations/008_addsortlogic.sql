-- +goose Up
ALTER TABLE event_groups
    ADD COLUMN sort_logic TEXT NOT NULL DEFAULT 'balanced'
    CHECK (sort_logic IN ('balanced', 'ranked'));

-- +goose Down
ALTER TABLE event_groups DROP COLUMN sort_logic;
