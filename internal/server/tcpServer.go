// Package server is the server
package server

import (
	"context"
	"log/slog"
	"net"

	"github.com/Arush71/redis-server/internal/storage"
)

type Server struct {
	logger     *slog.Logger
	listenAddr string
	ln         net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	storage    *storage.Storage
}

func NewServer(listenAddr string, logger *slog.Logger, storage *storage.Storage) *Server {
	s := &Server{
		listenAddr: listenAddr,
		logger:     logger,
		storage:    storage,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln
	go s.acceptLoop()
	<-s.ctx.Done()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			s.logger.Error("accept connection error", "err", err)
			continue
		}
		go s.manageConnection(conn)
	}
}
