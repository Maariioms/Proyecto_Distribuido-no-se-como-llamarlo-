package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "soc-mvp-distribuido/events/gen/eventspb/events/v1"
	"soc-mvp-distribuido/events/internal/model"
	"soc-mvp-distribuido/events/internal/repository"
)

// el "controller" de grpc de events
// igual que en accounts -> misma logica que el http y el MISMO repositorio
type EventServer struct {
	// pa que compile aunque falte algun rpc, igual que en account_server.go
	eventsv1.UnimplementedEventServiceServer

	repo repository.EventRepository
}

func NewEventServer(repo repository.EventRepository) *EventServer {
	return &EventServer{repo: repo}
}

func (s *EventServer) CreateEvent(ctx context.Context, req *eventsv1.CreateEventRequest) (*eventsv1.CreateEventResponse, error) {
	if req.GetType() == "" || req.GetSeverity() == "" {
		return nil, status.Error(codes.InvalidArgument, "type and severity are required")
	}

	event, err := s.repo.Create(ctx, req.GetType(), req.GetSeverity())
	if err != nil {
		return nil, status.Error(codes.Internal, "could not create event")
	}
	return &eventsv1.CreateEventResponse{Event: toProto(event)}, nil
}

func (s *EventServer) ListEvents(ctx context.Context, _ *eventsv1.ListEventsRequest) (*eventsv1.ListEventsResponse, error) {
	events, err := s.repo.List(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not list events")
	}

	// pasa cada model.Event a mensaje de grpc uno por uno
	out := make([]*eventsv1.Event, 0, len(events))
	for _, e := range events {
		out = append(out, toProto(e))
	}
	return &eventsv1.ListEventsResponse{Events: out}, nil
}

func (s *EventServer) GetEvent(ctx context.Context, req *eventsv1.GetEventRequest) (*eventsv1.GetEventResponse, error) {
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	event, err := s.repo.Get(ctx, uint(req.GetId()))
	if errors.Is(err, repository.ErrNotFound) {
		// NotFound -> el 404 de grpc
		return nil, status.Error(codes.NotFound, "event not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "could not fetch event")
	}
	return &eventsv1.GetEventResponse{Event: toProto(event)}, nil
}

// pasa de model.Event de go a eventsv1.Event que es el mensaje de red
func toProto(e model.Event) *eventsv1.Event {
	return &eventsv1.Event{
		Id:        uint64(e.ID),
		Type:      e.Type,
		Severity:  e.Severity,
		CreatedAt: timestamppb.New(e.CreatedAt),
	}
}
