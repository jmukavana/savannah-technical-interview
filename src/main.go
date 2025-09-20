package main

import (
	"context"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"savannah/src/Catalog"
	"savannah/src/Customer"
	"savannah/src/Logger"
	"savannah/src/Storage"
)

// @title           Catalog API
// @version         1.0
// @description     API documentation for Catalog service (products & categories).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	log := Logger.New()
	defer log.Sync()

	// DATABASE URL from env: POSTGRES_DSN e.g. postgres://user:pass@localhost:5432/dbname?sslmode=disable
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:1973@localhost:5432/savannah?sslmode=disable" // fallback DSN
	}

	db, err := Storage.NewPostgres(dsn)
	if err != nil {
		log.Fatal("db connect", zap.Error(err))
	}
	defer db.Close()

	// repos
	customerRepository := Customer.NewRepository(db, log)
	productRepository := Catalog.NewRepository(db, log)

	// services
	customerService := Customer.NewService(customerRepository, log)
	productService := Catalog.NewService(productRepository, log)

	// handler
	customerHandler := Customer.NewHandler(customerService, log)
	productHandler := Catalog.NewHandler(productService, log)

	r := chi.NewRouter()
	r.Use(Logger.ChiMiddleware(log))
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1/customers", func(r chi.Router) {
		r.Get("/", customerHandler.List)
		r.Post("/", customerHandler.Create)
		r.Get("/{id}", customerHandler.Get)
		r.Put("/{id}", customerHandler.Update)
		r.Delete("/{id}", customerHandler.Delete)
	})
	r.Route("/api/v1/categories", func(r chi.Router) {
		r.Post("/", productHandler.CreateCategory)
		r.Get("/{id}", productHandler.GetCategory)
	})
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Post("/", productHandler.CreateProduct)
		r.Get("/", productHandler.ListProducts)
		r.Get("/{id}", productHandler.GetCategory)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		log.Sugar().Infof("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	<-quit
	log.Sugar().Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Sugar().Info("server exiting")
}
