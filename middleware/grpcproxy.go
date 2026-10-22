package main

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	accountsv1 "soc-mvp-distribuido/middleware/gen/accountspb/accounts/v1"
	eventsv1 "soc-mvp-distribuido/middleware/gen/eventspb/events/v1"
	statusv1 "soc-mvp-distribuido/middleware/gen/statuspb/status/v1"
)

// este archivo es el load balancer pero pa grpc
//
// basicamente el middleware abre su propio puerto grpc que es el 9001
// y se hace pasar por los 3 servicios implementando AccountService EventService y StatusService
// llega una llamada -> elige una instancia sana de ese servicio con round-robin igual que el http
// y le hace LA MISMA llamada al backend
// osea es servidor pa quien llama y cliente pa el backend
//
// ojo que solo reenvia rpc unarios, osea una peticion y una respuesta
// que es lo unico que tienen nuestros .proto

// guarda una conexion grpc por direccion
// pa no abrir una nueva en cada llamada pq abrir es caro y grpc las reutiliza sin problema
type connPool struct {
	// aqui si va Mutex y no atomic
	// "revisar si existe y si no crearla" son varios pasos que pasan juntos
	// no es un solo numero
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

func newConnPool() *connPool {
	return &connPool{conns: make(map[string]*grpc.ClientConn)}
}

func (p *connPool) get(addr string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.conns[addr]; ok {
		return conn, nil
	}

	// NewClient no conecta de una vez y conecta en la primera llamada
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.conns[addr] = conn
	return conn, nil
}

// elige una instancia sana del servicio con el balancer, el mismo round-robin del http
// y regresa la conexion pa hablarle
func (p *connPool) pick(b *balancer, method string) (*grpc.ClientConn, error) {
	instance := b.pick()
	if instance == nil {
		// Unavailable -> el 503 de grpc
		return nil, status.Errorf(codes.Unavailable, "no hay instancias sanas de %s", b.service)
	}

	log.Printf("grpc dispatching %s -> %s", method, instance.Host)
	return p.get(instance.Host)
}

// un proxy por servicio
// los 3 hacen lo mismo -> elegir instancia y reenviar

type accountsProxy struct {
	accountsv1.UnimplementedAccountServiceServer
	pool     *connPool
	balancer *balancer
}

func (p *accountsProxy) Register(ctx context.Context, req *accountsv1.RegisterRequest) (*accountsv1.RegisterResponse, error) {
	conn, err := p.pool.pick(p.balancer, "accounts.v1.AccountService/Register")
	if err != nil {
		return nil, err
	}
	return accountsv1.NewAccountServiceClient(conn).Register(ctx, req)
}

func (p *accountsProxy) Login(ctx context.Context, req *accountsv1.LoginRequest) (*accountsv1.LoginResponse, error) {
	conn, err := p.pool.pick(p.balancer, "accounts.v1.AccountService/Login")
	if err != nil {
		return nil, err
	}
	return accountsv1.NewAccountServiceClient(conn).Login(ctx, req)
}

type eventsProxy struct {
	eventsv1.UnimplementedEventServiceServer
	pool     *connPool
	balancer *balancer
}

func (p *eventsProxy) CreateEvent(ctx context.Context, req *eventsv1.CreateEventRequest) (*eventsv1.CreateEventResponse, error) {
	conn, err := p.pool.pick(p.balancer, "events.v1.EventService/CreateEvent")
	if err != nil {
		return nil, err
	}
	return eventsv1.NewEventServiceClient(conn).CreateEvent(ctx, req)
}

func (p *eventsProxy) ListEvents(ctx context.Context, req *eventsv1.ListEventsRequest) (*eventsv1.ListEventsResponse, error) {
	conn, err := p.pool.pick(p.balancer, "events.v1.EventService/ListEvents")
	if err != nil {
		return nil, err
	}
	return eventsv1.NewEventServiceClient(conn).ListEvents(ctx, req)
}

func (p *eventsProxy) GetEvent(ctx context.Context, req *eventsv1.GetEventRequest) (*eventsv1.GetEventResponse, error) {
	conn, err := p.pool.pick(p.balancer, "events.v1.EventService/GetEvent")
	if err != nil {
		return nil, err
	}
	return eventsv1.NewEventServiceClient(conn).GetEvent(ctx, req)
}

type statusProxy struct {
	statusv1.UnimplementedStatusServiceServer
	pool     *connPool
	balancer *balancer
}

func (p *statusProxy) GetStatus(ctx context.Context, req *statusv1.GetStatusRequest) (*statusv1.GetStatusResponse, error) {
	conn, err := p.pool.pick(p.balancer, "status.v1.StatusService/GetStatus")
	if err != nil {
		return nil, err
	}
	return statusv1.NewStatusServiceClient(conn).GetStatus(ctx, req)
}

// abre el puerto grpc del middleware y deja el proxy corriendo en una goroutine
// igual que el servidor http y no bloquea el main
func startGRPCProxy(port string, accountsB, eventsB, statusB *balancer) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	pool := newConnPool()
	s := grpc.NewServer()
	accountsv1.RegisterAccountServiceServer(s, &accountsProxy{pool: pool, balancer: accountsB})
	eventsv1.RegisterEventServiceServer(s, &eventsProxy{pool: pool, balancer: eventsB})
	statusv1.RegisterStatusServiceServer(s, &statusProxy{pool: pool, balancer: statusB})

	// reflection pa que grpcurl le pueda preguntar al middleware que servicios tiene
	reflection.Register(s)

	go func() {
		log.Printf("middleware grpc listening on :%s", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()
}
