-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN user_id;
-- +goose StatementEnd
