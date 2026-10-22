package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	accountsv1 "soc-mvp-distribuido/backend/gen/accountspb/accounts/v1"
	"soc-mvp-distribuido/backend/internal/events"
	"soc-mvp-distribuido/backend/internal/model"
	"soc-mvp-distribuido/backend/internal/repository"
)

// el "controller" pero de grpc
// hace lo mismo que AccountController, el http, y usa el MISMO repositorio
// cambia el transporte y la logica es la misma y esa es la gracia del repository
type AccountServer struct {
	// obligatorio en grpc-go
	// trae un "no implementado" por default pa cada rpc
	// asi si agregamos un rpc al .proto y aqui no lo escribimos no truena la compilacion
	accountsv1.UnimplementedAccountServiceServer

	repo      repository.AccountRepository
	publisher events.Publisher
}

func NewAccountServer(repo repository.AccountRepository, publisher events.Publisher) *AccountServer {
	return &AccountServer{repo: repo, publisher: publisher}
}

func (s *AccountServer) Register(ctx context.Context, req *accountsv1.RegisterRequest) (*accountsv1.RegisterResponse, error) {
	// mismas reglas que el http
	if req.GetUsername() == "" || len(req.GetPassword()) < 6 {
		return nil, status.Error(codes.InvalidArgument, "username is required and password must be at least 6 characters")
	}

	account, err := s.repo.Create(ctx, req.GetUsername(), req.GetPassword())
	if errors.Is(err, repository.ErrUsernameTaken) {
		// AlreadyExists -> el 409 de grpc
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "could not create account")
	}

	s.publisher.Publish(ctx, "account_created", "low")
	return &accountsv1.RegisterResponse{Account: toProto(account)}, nil
}

func (s *AccountServer) Login(ctx context.Context, req *accountsv1.LoginRequest) (*accountsv1.LoginResponse, error) {
	account, err := s.repo.Authenticate(ctx, req.GetUsername(), req.GetPassword())
	if errors.Is(err, repository.ErrInvalidCredentials) {
		s.publisher.Publish(ctx, "login_failed", "medium")
		// Unauthenticated -> el 401 de grpc
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "could not authenticate")
	}

	s.publisher.Publish(ctx, "login_success", "low")
	return &accountsv1.LoginResponse{Account: toProto(account)}, nil
}

// pasa de model.Account de go a accountsv1.Account que es el mensaje de red
// ojo que NO copia el PasswordHash asi que nunca sale del servicio
func toProto(a model.Account) *accountsv1.Account {
	return &accountsv1.Account{
		Id:        uint64(a.ID),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}
