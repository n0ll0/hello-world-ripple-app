package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/model"
)

var ErrNotFound = errors.New("not found")

func (db *DB) CreateUser(ctx context.Context, user *model.User) error {
	res, err := db.ExecContext(ctx, `INSERT INTO users (username, password_hash) VALUES (?, ?)`, user.Username, user.PasswordHash)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	row := db.QueryRowContext(ctx, `SELECT created_at FROM users WHERE id = ?`, id)
	if err := row.Scan(&user.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := db.QueryRowContext(ctx, `SELECT id, username, password_hash, created_at FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) ListUsers(ctx context.Context) ([]model.User, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, username, password_hash, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
