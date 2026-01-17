-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS link_visits
(
    id         BIGSERIAL PRIMARY KEY,
    link_id    BIGINT      NOT NULL REFERENCES links (id) ON DELETE CASCADE,
    ip         VARCHAR(40) NOT NULL,
    user_agent TEXT        NOT NULL,
    referer    TEXT,
    status     INT         NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link_visits;
-- +goose StatementEnd
