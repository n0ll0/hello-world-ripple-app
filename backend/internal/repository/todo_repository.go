package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/model"
)

func (db *DB) ListTodosByUser(ctx context.Context, userID int64) ([]model.Todo, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, user_id, title, completed, created_at FROM todos WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, *todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

func (db *DB) CreateTodo(ctx context.Context, todo *model.Todo) error {
	now := time.Now().UTC()
	res, err := db.ExecContext(ctx, `INSERT INTO todos (user_id, title, completed, created_at) VALUES (?, ?, ?, ?)`, todo.UserID, todo.Title, boolToInt(todo.Completed), now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = id
	todo.CreatedAt = now
	return nil
}

func (db *DB) UpdateTodo(ctx context.Context, todo *model.Todo) error {
	_, err := db.ExecContext(ctx, `UPDATE todos SET title = ?, completed = ? WHERE id = ? AND user_id = ?`, todo.Title, boolToInt(todo.Completed), todo.ID, todo.UserID)
	return err
}

func (db *DB) DeleteTodo(ctx context.Context, userID, todoID int64) error {
	res, err := db.ExecContext(ctx, `DELETE FROM todos WHERE id = ? AND user_id = ?`, todoID, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *DB) GetTodo(ctx context.Context, userID, todoID int64) (*model.Todo, error) {
	row := db.QueryRowContext(ctx, `SELECT id, user_id, title, completed, created_at FROM todos WHERE id = ? AND user_id = ?`, todoID, userID)
	return scanTodo(row)
}

type todoScanner interface {
	Scan(dest ...any) error
}

func scanTodo(scanner todoScanner) (*model.Todo, error) {
	var (
		t         model.Todo
		completed int
		createdAt any
	)
	if err := scanner.Scan(&t.ID, &t.UserID, &t.Title, &completed, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t.Completed = completed == 1
	switch v := createdAt.(type) {
	case time.Time:
		t.CreatedAt = v
	case string:
		if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			t.CreatedAt = parsed
		} else if parsed, err := time.Parse(time.RFC3339Nano, v); err == nil {
			t.CreatedAt = parsed
		}
	case []byte:
		if parsed, err := time.Parse("2006-01-02 15:04:05", string(v)); err == nil {
			t.CreatedAt = parsed
		}
	default:
		// leave zero value
	}
	return &t, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
