-- +goose Up
INSERT OR IGNORE INTO users (id, username, password_hash)
VALUES (1, 'demo', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAg9q4CP.et8DmUbEQqcWgQBayg42');

INSERT OR IGNORE INTO todos (user_id, title, completed)
VALUES (1, 'Try the new Ripple app backend', 0);

-- +goose Down
DELETE FROM todos;
DELETE FROM users;
