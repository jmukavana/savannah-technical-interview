package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"savannah/src/Billing"
	"savannah/src/Catalog"
	"savannah/src/Customer"
	"savannah/src/Inventory"
	"savannah/src/Logger"
	"savannah/src/Orders"
	"savannah/src/Storage"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// main is the entry point for the program.
//
// It sets up a logger, opens a Postgres database connection,
// creates a customer repository and service, and sets up a
// HTTP handler using Chi.
//
// It also sets up a graceful shutdown mechanism using a
// context and a quit channel.
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
	catRepo := Catalog.NewRepository(db, log)
	invRepo := Inventory.NewRepository(db, log)
	ordersRepo := Orders.NewRepository(db, log)
	billingRepo := Billing.NewRepository(db, log)
	customerRepo := Customer.NewRepository(db, log)

	// services
	invSvc := Inventory.NewService(invRepo, db, log)
	ordersSvc := Orders.NewService(ordersRepo, db, invSvc, log)
	no := &billing.NoopProvider{}
	billingSvc := Billing.NewService(billingRepo, no, log)
	customerSvc := Customer.NewService(customerRepo, log)
	h := Customer.NewHandler(customerSvc, log)

	r := chi.NewRouter()
	r.Use(Logger.ChiMiddleware(log))

	r.Route("/api/v1/customers", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", ordersSvc.Create))
		r.Get("/", ordersSvc.List)
		r.Get("/{id}", ordersSvc.Get)
		r.Put("/{id}", ordersSvc.Update)
		r.Delete("/{id}", ordersSvc.Delete)
	})
	r.Route("/api/v1/invoices", func(r chi.Router) {
		r.Post("/", billingSvc.Create)
		r.Get("/", billingSvc.List)
		r.Get("/{id}", billingSvc.Get)
		r.Put("/{id}", billingSvc.Update)
		r.Delete("/{id}", billingSvc.Delete)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		log.Sugar().Infof("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	<-quit
	log.Sugar().Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Sugar().Info("server exiting")
}
