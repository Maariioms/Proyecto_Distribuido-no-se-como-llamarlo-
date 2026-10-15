package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"soc-mvp-distribuido/events/internal/consul"
	"soc-mvp-distribuido/events/internal/controller"
	"soc-mvp-distribuido/events/internal/db"
	"soc-mvp-distribuido/events/internal/repository"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=soc password=soc dbname=soc_events port=5432 sslmode=disable"
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	eventRepo := repository.NewGormEventRepository(gormDB)
	eventController := controller.NewEventController(eventRepo)

	controller.StartGenerator(eventRepo, 15*time.Second)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

	r.Get("/health", controller.Health)
	r.Route("/api/events", func(r chi.Router) {
		r.Post("/", eventController.Create)
		r.Get("/", eventController.List)
		r.Get("/{id}", eventController.Get)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Consul va a llamar a /health cada 5s para saber si seguimos vivos.
	if err := consul.Register("/health"); err != nil {
		log.Printf("WARNING: %v", err)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	// Docker manda SIGTERM cuando hacemos "docker stop". Al recibirlo nos
	// damos de baja de Consul y luego apagamos el servidor sin cortar nada.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		consul.Deregister()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("[%s] listening on :%s", os.Getenv("SERVICE_ID"), port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
