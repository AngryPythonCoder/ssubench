package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ssubench/internal/config"
	"ssubench/internal/domain"
	"ssubench/internal/handler"
	"ssubench/internal/middleware"
	"ssubench/internal/repository"
	"ssubench/internal/service"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

func main() {
	var exitCode int
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	cfg, err := config.Load()

	ctx := context.Background()

	db, err := repository.NewPGXConnection(ctx, cfg.DBURL)
	if err != nil {
		log.Println(err)
		exitCode = 1
		return
	}

	defer db.Pool.Close()
	log.Println("DB connection is estabilished")

	txManager := repository.NewPGXTXManager(db.Pool)
	userRepo := repository.NewPGXUserRepository(db.Pool)
	taskRepo := repository.NewPGXTaskRepository(db.Pool)
	paymentRepo := repository.NewPGXPaymentRepository(db.Pool)

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
	r.Use(
		chimiddleware.Timeout(60*time.Second),
		chimiddleware.RequestID,
		chimiddleware.Logger,
		chimiddleware.Recoverer,
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

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server is running on port " + cfg.ServerPort)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v\n", err)
		}
	}()

	<-stop
	log.Println("Shutting down the server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Server cannot shut down correctly: %v\n", err)
		exitCode = 1
		return
	}

	log.Println("Server stopped")
}
