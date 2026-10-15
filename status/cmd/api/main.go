package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"soc-mvp-distribuido/status/internal/consul"
	"soc-mvp-distribuido/status/internal/controller"
)

func main() {
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./status.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("opening sqlite: %v", err)
	}
	defer db.Close()

	if err := controller.InitTracker(db); err != nil {
		log.Fatalf("init tracker: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

	r.Get("/health", controller.Status)
	r.Get("/heartbeat", controller.Heartbeat)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Consul revisa /heartbeat (no /health) cada 5s para saber si seguimos vivos.
	if err := consul.Register("/heartbeat"); err != nil {
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
