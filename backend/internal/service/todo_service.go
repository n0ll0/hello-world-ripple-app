package service

import (
	"context"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/model"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/repository"
)

func NewTodoService(db *repository.DB) *TodoService {
	return &TodoService{db: db}
}

type TodoService struct {
	db *repository.DB
}

func (s *TodoService) ListByUser(ctx context.Context, userID int64) ([]model.Todo, error) {
	return s.db.ListTodosByUser(ctx, userID)
}

func (s *TodoService) Create(ctx context.Context, todo *model.Todo) error {
	return s.db.CreateTodo(ctx, todo)
}

func (s *TodoService) Update(ctx context.Context, todo *model.Todo) error {
	return s.db.UpdateTodo(ctx, todo)
}

func (s *TodoService) Delete(ctx context.Context, userID, todoID int64) error {
	return s.db.DeleteTodo(ctx, userID, todoID)
}

func (s *TodoService) Get(ctx context.Context, userID, todoID int64) (*model.Todo, error) {
	return s.db.GetTodo(ctx, userID, todoID)
}
