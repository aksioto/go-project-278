-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_links_short_name_include_url
    ON links(short_name)
    INCLUDE(original_url, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_links_short_name_include_url;
-- +goose StatementEnd