CREATE TABLE logs.goods 
(
    id Int32,
    project_id Int32,
    name String,
    description String,
    priority Int64,
    removed Bool,
    created_at DateTime,
    event_time DateTime
)
ENGINE = MergeTree()
PRIMARY KEY (id, project_id, name);