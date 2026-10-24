package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	statusv1 "soc-mvp-distribuido/status/gen/statuspb/status/v1"
	"soc-mvp-distribuido/status/internal/consul"
	"soc-mvp-distribuido/status/internal/consumer"
	"soc-mvp-distribuido/status/internal/controller"
	"soc-mvp-distribuido/status/internal/grpcserver"
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

	// ---- kafka ----
	// si hay KAFKA_BROKERS arranca el consumidor que cuenta los eventos que va guardando events
	appCtx, stopApp := context.WithCancel(context.Background())
	defer stopApp()
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		consumer.StartEventsConsumer(appCtx, brokers)
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

	// ---- grpc ----
	// ahora el servicio tiene DOS puertos -> el http de siempre en el 8080 y este grpc
	// los dos usan el mismo Tracker que es el Singleton y solo cambia como llegan las peticiones
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcSrv := grpcserver.New()
	statusv1.RegisterStatusServiceServer(grpcSrv, grpcserver.NewStatusServer())
	go func() {
		log.Printf("[%s] grpc listening on :%s", os.Getenv("SERVICE_ID"), grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// consul revisa /heartbeat y no /health cada 5s por http y el health de grpc por el otro puerto
	// son dos registros -> status y status-grpc
	if err := consul.Register("/heartbeat"); err != nil {
		log.Printf("WARNING: %v", err)
	}
	if err := consul.RegisterGRPC(); err != nil {
		log.Printf("WARNING: %v", err)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	// docker manda SIGTERM con "docker stop"
	// al recibirlo nos damos de baja de consul y apagamos los dos servidores sin cortar nada
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		// cancela el ctx y el consumidor de kafka termina
		stopApp()
		consul.Deregister()
		grpcSrv.GracefulStop()
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
