CREATE TYPE provider AS ENUM(
    'fields',
    'yandex'
);

ALTER TABLE users ADD COLUMN provider provider;