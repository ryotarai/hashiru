package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ryotarai/hashiru/gen/hashiru/v1/hashiruv1connect"
	"github.com/ryotarai/hashiru/server/command"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Run starts the command execution server.
func Run(ctx context.Context, socketPath string) error {
	command := &command.CommandServer{}
	mux := http.NewServeMux()
	path, handler := hashiruv1connect.NewCommandServiceHandler(command)
	mux.Handle(path, handler)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer os.Remove(socketPath)
	if err := os.Chmod(socketPath, 0700); err != nil {
		return err
	}

	server := &http.Server{
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		server.Shutdown(ctx)
		listener.Close()
	}()

	log.Printf("Starting server on %s", socketPath)
	return server.Serve(listener)
}
