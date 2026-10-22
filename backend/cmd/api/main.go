package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	accountsv1 "soc-mvp-distribuido/backend/gen/accountspb/accounts/v1"
	"soc-mvp-distribuido/backend/internal/consul"
	"soc-mvp-distribuido/backend/internal/controller"
	"soc-mvp-distribuido/backend/internal/db"
	"soc-mvp-distribuido/backend/internal/events"
	"soc-mvp-distribuido/backend/internal/grpcserver"
	"soc-mvp-distribuido/backend/internal/repository"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=soc password=soc dbname=soc port=5432 sslmode=disable"
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	accountRepo := repository.NewGormAccountRepository(gormDB)

	// a quien le avisamos cuando hay login o registro
	// si no hay EVENTS_GRPC_ADDR no avisamos a nadie y usa Noop pa que todo siga igual
	var publisher events.Publisher = events.Noop{}
	if addr := os.Getenv("EVENTS_GRPC_ADDR"); addr != "" {
		p, err := events.NewGRPCPublisher(addr)
		if err != nil {
			log.Printf("WARNING: no se pudo armar el publisher de events: %v", err)
		} else {
			publisher = p
		}
	}

	accountController := controller.NewAccountController(accountRepo, publisher)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

	r.Get("/health", controller.Health)
	r.Route("/api/accounts", func(r chi.Router) {
		r.Post("/register", accountController.Register)
		r.Post("/login", accountController.Login)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ---- grpc ----
	// ahora el servicio tiene DOS puertos -> el http de siempre en el 8080 y este grpc
	// los dos usan el mismo repositorio y solo cambia como llegan las peticiones
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcSrv := grpcserver.New()
	accountsv1.RegisterAccountServiceServer(grpcSrv, grpcserver.NewAccountServer(accountRepo, publisher))
	go func() {
		log.Printf("[%s] grpc listening on :%s", os.Getenv("SERVICE_ID"), grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// consul le pega a /health por http y al health de grpc cada 5s pa ver si seguimos vivos
	// son dos registros -> accounts y accounts-grpc
	if err := consul.Register("/health"); err != nil {
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
