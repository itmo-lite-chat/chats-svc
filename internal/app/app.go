package app

import (
	"context"
	"errors"
	"net"

	"github.com/itmo-lite-chat/chats-svc/cmd/config"
	"github.com/itmo-lite-chat/chats-svc/internal/api/grpc_api"
	"github.com/itmo-lite-chat/chats-svc/internal/service"
	"github.com/itmo-lite-chat/chats-svc/internal/storage"
	pb "github.com/itmo-lite-chat/proto-registry/gen/services/chats_service/chats/v1"
	"google.golang.org/grpc"
)

type App struct {
	cfg     config.Config
	storage *storage.Storage
	server  *grpc.Server
}

func New(cfg config.Config) (*App, error) {
	return &App{cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context) error {
	st, err := storage.New(ctx, a.cfg.PostgresDSN)
	if err != nil {
		return err
	}
	a.storage = st

	api := grpc_api.New(service.New(st))

	listener, err := net.Listen("tcp", a.cfg.GRPCAddr)
	if err != nil {
		return err
	}

	a.server = grpc.NewServer()
	pb.RegisterChatsServiceServer(a.server, api)

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		a.server.GracefulStop()
		a.storage.Close()
		return nil
	case err := <-errCh:
		a.storage.Close()
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
}
