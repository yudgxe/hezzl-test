-- +goose Up
-- +goose StatementBegin
INSERT INTO projects(id, name) VALUES (1, 'test');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM projects WHERE id = 1;
-- +goose StatementEnd
