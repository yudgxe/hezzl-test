-- Создание товара
-- name: CreateGood :one
INSERT INTO goods (name, project_id) VALUES (@name::varchar, @project_id::int) RETURNING *;

-- Обновление товара.
-- name: UpdateGood :one
UPDATE goods SET
    name = @name,
    description = coalesce(sqlc.narg(description), description)
WHERE id = @id AND project_id = @project_id
RETURNING *;

-- Обновление статуса удаления товара.
-- name: UpdateGoodRemoved :one
UPDATE goods SET removed = @removed WHERE id = @id AND project_id = @project_id RETURNING *;

-- Список всех товаров.
-- name: ListGoods :many
SELECT * FROM goods
ORDER BY id
LIMIT $1 OFFSET $2;

-- Метаданные в частности, кол-во записей и кол-во удаленных записей.
-- name: MetaGood :one
SELECT Count(*)::int as total, Count(*) FILTER(WHERE removed = TRUE)::int as removed FROM goods;

-- Существует ли товар.
-- name: HasGood :one
SELECT EXISTS (SELECT 1 FROM goods WHERE id = @id AND project_id = @project_id AND removed = FALSE LIMIT 1);

-- Пересчет преоритетов товара.
-- name: ReprioritiizeGood :many
WITH old AS (SELECT priority FROM goods where id = @id AND goods.project_id = @project_id)
UPDATE goods SET priority =
    CASE
        WHEN goods.id = @id THEN @priority
        WHEN goods.id <> @id AND ((select priority from old) >= @priority AND priority >= @priority AND priority < (select priority from old)) OR ((select priority from old) < @priority AND priority >= @priority) THEN priority+1
        ELSE priority
    END
WHERE goods.id = @id OR (goods.id <> @id  AND (select priority from old) >= @priority AND priority >= @priority AND priority < (select priority from old) OR ((select priority from old) < @priority AND priority >= @priority)) RETURNING *;
