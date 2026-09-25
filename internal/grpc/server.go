package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"

	spb "sparkbridge/pkg/pb"

	"google.golang.org/grpc"
)

type BridgeManager interface {
	Publish(context.Context, *spb.EdgeMessage) error
	Subscribe() <-chan *spb.CloudCommand
}

type Server struct {
	cfg      Config
	log      *slog.Logger
	mgr      BridgeManager
	grpc     *grpc.Server
	listener net.Listener
	done     chan error
	mu       sync.Mutex
}

type Config struct {
	SocketPath        string
	SocketPermissions int
}

func NewServer(cfg Config, mgr BridgeManager, log *slog.Logger) (*Server, error) {
	return &Server{cfg: cfg, log: log, mgr: mgr, done: make(chan error, 1)}, nil
}

func (s *Server) Start() error {
	if err := os.MkdirAll(filepath.Dir(s.cfg.SocketPath), 0o755); err != nil {
		return fmt.Errorf("mkdir socket dir: %w", err)
	}
	_ = os.Remove(s.cfg.SocketPath)
	ln, err := net.Listen("unix", s.cfg.SocketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket: %w", err)
	}
	if err := os.Chmod(s.cfg.SocketPath, os.FileMode(s.cfg.SocketPermissions)); err != nil {
		_ = ln.Close()
		return fmt.Errorf("chmod socket: %w", err)
	}
	s.listener = ln
	s.grpc = grpc.NewServer()
	spb.RegisterSparkplugIngressServiceServer(s.grpc, &bridgeService{mgr: s.mgr})
	go func() { s.done <- s.grpc.Serve(ln) }()
	return nil
}

func (s *Server) Wait() error { return <-s.done }

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grpc != nil {
		s.grpc.GracefulStop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
		_ = os.Remove(s.cfg.SocketPath)
	}
	return nil
}

type bridgeService struct {
	spb.UnimplementedSparkplugIngressServiceServer
	mgr BridgeManager
}

func (b *bridgeService) Bridge(stream spb.SparkplugIngressService_BridgeServer) error {
	ctx := stream.Context()
	commands := b.mgr.Subscribe()
	sendErr := make(chan error, 1)
	recvErr := make(chan error, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				sendErr <- ctx.Err()
				return
			case cmd, ok := <-commands:
				if !ok || cmd == nil {
					sendErr <- nil
					return
				}
				if err := stream.Send(cmd); err != nil {
					sendErr <- err
					return
				}
			}
		}
	}()

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				recvErr <- err
				return
			}
			if err := b.mgr.Publish(ctx, msg); err != nil {
				recvErr <- err
				return
			}
		}
	}()

	select {
	case err := <-sendErr:
		return err
	case err := <-recvErr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
