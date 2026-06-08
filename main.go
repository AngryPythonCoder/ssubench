package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ssubench/internal/app"
	"ssubench/internal/config"
	"ssubench/internal/repository"
	"syscall"
	"time"
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

	router := app.SetupHandler(db.Pool, cfg, true)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
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
