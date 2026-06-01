package app

import (
	"net/http"
	"ssubench/internal/config"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
	"ssubench/internal/middleware"
	"ssubench/internal/repository"
	"ssubench/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupHandler(pool *pgxpool.Pool, cfg *config.Config, useLogger bool) http.Handler {
	txManager := repository.NewPGXTXManager(pool)
	userRepo := repository.NewPGXUserRepository(pool)
	taskRepo := repository.NewPGXTaskRepository(pool)
	paymentRepo := repository.NewPGXPaymentRepository(pool)

	authService := service.NewAuthService(txManager, userRepo, cfg)
	userService := service.NewUserService(txManager, userRepo)
	taskService := service.NewTaskService(txManager, taskRepo, userRepo, paymentRepo)
	paymentService := service.NewPaymentService(paymentRepo)

	validate := validator.New(validator.WithRequiredStructEnabled())

	authHandler := handler.NewAuthHandler(authService, validate)
	userHandler := handler.NewUserHandler(userService, validate, cfg.MaxPaginationLimit)
	taskHandler := handler.NewTaskHandler(taskService, validate, cfg.MaxPaginationLimit)
	paymentHandler := handler.NewPaymentHandler(paymentService, validate, cfg.MaxPaginationLimit)

	adminChecker := middleware.RoleChecker(func(role domain.UserRole) bool {
		return role == domain.RoleAdmin
	})
	customerChecker := middleware.RoleChecker(func(role domain.UserRole) bool {
		return role == domain.RoleCustomer || role == domain.RoleAdmin
	})
	performerChecker := middleware.RoleChecker(func(role domain.UserRole) bool {
		return role == domain.RolePerformer || role == domain.RoleAdmin
	})

	authorizer := middleware.Auth(cfg.JWTSecret)
	blockChecker := middleware.BlockChecker(userService)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	if useLogger {
		r.Use(chimiddleware.Logger)
	}
	r.Use(
		chimiddleware.Recoverer,
		middleware.StandardTimeout(60*time.Second, "Время вышло"),
	)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(
			authorizer,
			blockChecker,
		)

		r.Get("/auth/dumbcheck", authHandler.DumbCheck)
		r.Get("/users", userHandler.List)
		r.Get("/users/{user_id}", userHandler.Get)
		r.Get("/tasks", taskHandler.List)
		r.Get("/tasks/{task_id}", taskHandler.Get)
		r.Get("/tasks/{task_id}/bids", taskHandler.ListBids)
		r.Get("/tasks/{task_id}/bids/{bid_id}", taskHandler.GetBid)
		r.Get("/payments", paymentHandler.List)
		r.Get("/payments/{payment_id}", paymentHandler.Get)

		r.Group(func(r chi.Router) {
			r.Use(adminChecker)

			r.Post("/users/{user_id}/block", userHandler.Block)
			r.Post("/users/{user_id}/unblock", userHandler.Unblock)
			r.Post("/users/{user_id}/set-balance", userHandler.SetBalance)
		})

		r.Group(func(r chi.Router) {
			r.Use(customerChecker)

			r.Post("/tasks", taskHandler.Create)
			r.Post("/tasks/{task_id}/bids/{bid_id}/accept", taskHandler.AcceptBid)
			r.Post("/tasks/{task_id}/confirm", taskHandler.Complete)
			r.Post("/tasks/{task_id}/cancel", taskHandler.Cancel)
		})

		r.Group(func(r chi.Router) {
			r.Use(performerChecker)

			r.Post("/tasks/{task_id}/respond", taskHandler.CreateBid)
			r.Post("/tasks/{task_id}/finish", taskHandler.Finish)
		})
	})

	return r
}
