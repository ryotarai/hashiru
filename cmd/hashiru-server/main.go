package main

import (
	"context"
	"flag"
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

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	socketPath := flag.String("socket", "/tmp/hashiru.sock", "socket path")
	flag.Parse()

	command := &command.CommandServer{}
	mux := http.NewServeMux()
	path, handler := hashiruv1connect.NewCommandServiceHandler(command)
	mux.Handle(path, handler)

	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer os.Remove(*socketPath)
	if err := os.Chmod(*socketPath, 0700); err != nil {
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

	return server.Serve(listener)
}
