package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"

	"savannah/src/Catalog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()

	defer logger.Sync()

	db, err := sqlx.Connect("postgres", "postgres://postgres:1973@localhost:5432/savannah?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize layers
	catalogRepository := Catalog.NewRepository(db, logger)
	catalogService := Catalog.NewService(catalogRepository, logger)
	catalogHandler := Catalog.NewHandler(catalogService, logger)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1/", func(r chi.Router) {
		catalogHandler.RegisterRoutes(r)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		logger.Sugar().Infof("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	<-quit
	logger.Sugar().Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Sugar().Info("server exiting")
}
