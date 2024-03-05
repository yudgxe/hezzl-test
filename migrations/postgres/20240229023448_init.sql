-- +goose Up
-- +goose StatementBegin

-- Проекты.
CREATE TABLE projects (
    id serial PRIMARY KEY,
    name varchar(255) NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ix_projects_id ON projects (id);

-- Товары.
CREATE TABLE goods (
    id serial,
    project_id integer REFERENCES projects (id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    -- Всега равен max(priority) + 1
    priority serial NOT NULL,
    removed boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, project_id, name)
);

CREATE INDEX ix_goods_id_project_id ON goods (id, project_id);

CREATE FUNCTION set_priority() RETURNS TRIGGER AS $$
    BEGIN
        NEW.priority = 1 + COALESCE((SELECT MAX(priority) FROM goods), 0);
        RETURN NEW;
    END
$$ LANGUAGE plpgsql;

CREATE TRIGGER goods_set_priority
BEFORE INSERT ON goods
FOR EACH ROW EXECUTE FUNCTION set_priority();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS goods_set_priority ON goods;
DROP FUNCTION IF EXISTS set_priority();
DROP INDEX IF EXISTS ix_goods_id_project_id;
DROP INDEX IF EXISTS ix_projects_id;
DROP TABLE IF EXISTS goods;
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd 
