package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arush71/redis-server/internal/server"
	"github.com/Arush71/redis-server/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	storage := storage.InitStorage()
	server := server.NewServer(":6379", logger, storage)
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("server failed to connect or crashed", "err", err)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
