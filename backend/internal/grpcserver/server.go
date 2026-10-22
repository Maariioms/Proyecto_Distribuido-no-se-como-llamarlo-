package grpcserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// arma el servidor grpc con lo que TODOS los backends necesitan
//   - health con el grpc.health.v1 y ahi consul pregunta "estas vivo?"
//   - reflection que deja a grpcurl ver los servicios sin tener el .proto
//
// despues cada servicio le registra el suyo, ej AccountService
func New() *grpc.Server {
	s := grpc.NewServer()
	healthpb.RegisterHealthServer(s, health.NewServer())
	reflection.Register(s)
	return s
}
