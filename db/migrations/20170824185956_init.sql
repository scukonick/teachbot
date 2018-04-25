-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE "users" (
  id         SERIAL PRIMARY KEY,
  first_name TEXT,
  last_name  TEXT,
  username   TEXT,
  tg_id      BIGINT,
  created_at TIMESTAMP
);

CREATE UNIQUE INDEX users_tg_id_idx
  ON users (tg_id);
CREATE UNIQUE INDEX users_username_idx
  ON users (username);

CREATE TABLE "tasks" (
  id   SERIAL PRIMARY KEY,
  task TEXT,
  image TEXT
);

CREATE TABLE "invalid_messages" (
  id         SERIAL PRIMARY KEY,
  user_id    INT REFERENCES users (id),
  message    TEXT,
  created_at TIMESTAMP
);

CREATE TABLE "user_tasks" (
  id           SERIAL PRIMARY KEY,
  task_id INT,
  user_id      INT,
  created_at   TIMESTAMP
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "user_tasks";
DROP TABLE "invalid_messages";
DROP TABLE "tasks";
DROP TABLE "users";
