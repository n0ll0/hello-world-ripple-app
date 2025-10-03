package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"github.com/go-chi/chi/v5"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/middleware"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/model"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/repository"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/service"
)

type TodoHandler struct {
	Todos      *service.TodoService
	CreatedHub *EventHub
	UpdatedHub *EventHub
	DeletedHub *EventHub
}

func NewTodoHandler(todos *service.TodoService) *TodoHandler {
	return &TodoHandler{Todos: todos}
}

func (h *TodoHandler) SetWebSocketHubs(created, updated, deleted *EventHub) {
	h.CreatedHub = created
	h.UpdatedHub = updated
	h.DeletedHub = deleted
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	todos, err := h.Todos.ListByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(todos)
}

type createTodoRequest struct {
	Title string `json:"title"`
}

type updateTodoRequest struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	case <-time.After(5 * time.Second):
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	todo := &model.Todo{UserID: userID, Title: req.Title}
	if err := h.Todos.Create(r.Context(), todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the created todo to all WebSocket clients
	if h.CreatedHub != nil {
		if data, err := json.Marshal(todo); err == nil {
			h.CreatedHub.Broadcast(data)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Simulate a 2 second delay
	select {
	case <-r.Context().Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	case <-time.After(5 * time.Second):
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	todoID, err := parseIDParam(r, "id")
	if err != nil {
		http.Error(w, "invalid todo id", http.StatusBadRequest)
		return
	}
	todo, err := h.Todos.Get(r.Context(), userID, todoID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req updateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	if err := h.Todos.Update(r.Context(), todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the updated todo to all WebSocket clients
	if h.UpdatedHub != nil {
		if data, err := json.Marshal(todo); err == nil {
			h.UpdatedHub.Broadcast(data)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	todoID, err := parseIDParam(r, "id")
	if err != nil {
		http.Error(w, "invalid todo id", http.StatusBadRequest)
		return
	}
	if err := h.Todos.Delete(r.Context(), userID, todoID); err != nil {
		if err == repository.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the deleted todo ID to all WebSocket clients
	if h.DeletedHub != nil {
		type deletedEvent struct {
			ID int64 `json:"id"`
		}
		if data, err := json.Marshal(deletedEvent{ID: todoID}); err == nil {
			h.DeletedHub.Broadcast(data)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDParam(r *http.Request, name string) (int64, error) {
	raw := chi.URLParam(r, name)
	return strconv.ParseInt(raw, 10, 64)
}
