-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_link_visits_link_id ON link_visits (link_id);
CREATE INDEX IF NOT EXISTS idx_link_visits_created_at ON link_visits (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_link_visits_created_at;
DROP INDEX IF EXISTS idx_link_visits_link_id;
-- +goose StatementEnd
