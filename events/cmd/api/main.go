package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

	log.Printf("[%s] listening on :%s", os.Getenv("SERVICE_ID"), port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
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
