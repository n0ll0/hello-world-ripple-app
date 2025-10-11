package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"

	"github.com/n0ll0/hello-world-ripple-app-backend/internal/config"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/handler"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/middleware"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/migrations"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/repository"
	"github.com/n0ll0/hello-world-ripple-app-backend/internal/service"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_foreign_keys=on&_busy_timeout=5000", cfg.DBPath)

	db, err := repository.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := migrations.Up(db.DB); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	userService := service.NewUserService(db)
	todoService := service.NewTodoService(db)

	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientStore := store.NewClientStore()
	for _, client := range cfg.Clients {
		clientStore.Set(client.ID, &models.Client{
			ID:     client.ID,
			Secret: client.Secret,
			Domain: client.Domain,
		})
	}
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (string, error) {
		user, err := userService.Authenticate(ctx, username, password)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(user.ID, 10), nil
	})

	oauthHandler := &handler.OAuthHandler{Srv: srv}
	oauthHandler.SetErrorHandlers()

	userHandler := handler.NewUserHandler(userService)
	todoHandler := handler.NewTodoHandler(todoService)

	// Initialize WebSocket event hubs - one per event type
	todoCreatedHub := handler.NewEventHub("todo:created")
	todoUpdatedHub := handler.NewEventHub("todo:updated")
	todoDeletedHub := handler.NewEventHub("todo:deleted")

	go todoCreatedHub.Run()
	go todoUpdatedHub.Run()
	go todoDeletedHub.Run()

	// Connect the hubs to the todo handler
	todoHandler.SetWebSocketHubs(todoCreatedHub, todoUpdatedHub, todoDeletedHub)

	r := chi.NewRouter()

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/health", handler.Health)
	r.Get("/authorize", oauthHandler.Authorize)
	r.Post("/token", oauthHandler.Token)

	r.Route("/api", func(api chi.Router) {
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("Request: %s %s", r.Method, r.URL.Path)
				next.ServeHTTP(w, r)
			})
		})
		api.Post("/users", userHandler.Register)

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.OAuth2Guard(srv))
			protected.Get("/users", userHandler.List)

			protected.Get("/todos", todoHandler.List)
			protected.Post("/todos", todoHandler.Create)
			protected.Put("/todos/{id}", todoHandler.Update)
			protected.Delete("/todos/{id}", todoHandler.Delete)
		})
	})

	// WebSocket endpoints - one per event type
	r.Get("/ws/todos/created", todoCreatedHub.HandleWebSocket)
	r.Get("/ws/todos/updated", todoUpdatedHub.HandleWebSocket)
	r.Get("/ws/todos/deleted", todoDeletedHub.HandleWebSocket)

	log.Printf("Starting server with OAuth2 and SQLite on :%s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
