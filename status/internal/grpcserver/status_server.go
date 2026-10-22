package grpcserver

import (
	"context"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	statusv1 "soc-mvp-distribuido/status/gen/statuspb/status/v1"
	"soc-mvp-distribuido/status/internal/controller"
)

// el servidor grpc de status
// usa el mismo Tracker que el http y ese es el Singleton
// por eso llama a controller.CheckAndSnapshot y no tiene logica propia
type StatusServer struct {
	statusv1.UnimplementedStatusServiceServer
}

func NewStatusServer() *StatusServer {
	return &StatusServer{}
}

func (s *StatusServer) GetStatus(_ context.Context, _ *statusv1.GetStatusRequest) (*statusv1.GetStatusResponse, error) {
	uptime, checks, err := controller.CheckAndSnapshot()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not register check")
	}

	return &statusv1.GetStatusResponse{
		Status:     "ok",
		ServiceId:  os.Getenv("SERVICE_ID"),
		Uptime:     uptime.String(),
		ChecksDone: int64(checks),
	}, nil
}
