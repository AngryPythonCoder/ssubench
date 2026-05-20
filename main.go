package main

import (
	"context"
	"log"
	"net/http"
	"ssubench/internal/config"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
	"ssubench/internal/middleware"
	"ssubench/internal/repository"
	"ssubench/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

func main() {
	cfg, err := config.Load()

	ctx := context.Background()

	db, err := repository.NewPostgresConnection(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Pool.Close()
	log.Println("DB connection is estabilished")

	userRepo := repository.NewUserRepository(db.Pool)

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg)

	validate := validator.New(validator.WithRequiredStructEnabled())

	authHandler := handler.NewAuthHandler(authService, validate)
	userHandler := handler.NewUserHandler(userService, validate, cfg.MaxPaginationLimit)

	adminChecker := middleware.RoleChecker(domain.RoleAdmin)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))

		r.Get("/auth/dumbcheck", authHandler.DumbCheck)
		r.Get("/users", userHandler.List)
		r.Get("/users/{user_id}", userHandler.Get)

		r.Group(func(r chi.Router) {
			r.Use(adminChecker)

			r.Post("/users/{user_id}/block", userHandler.Block)
			r.Post("/users/{user_id}/unblock", userHandler.Unblock)
			r.Post("/users/{user_id}/set-balance", userHandler.SetBalance)
		})
	})

	log.Println("Running on port " + cfg.ServerPort)
	http.ListenAndServe(":"+cfg.ServerPort, r)
}
