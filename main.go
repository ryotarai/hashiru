package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/ryotarai/hashiru/cli/client"
	"github.com/ryotarai/hashiru/cli/server"
)

const defaultSocketPath = "/tmp/hashiru.sock"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify a command: server or client")
	}

	command := os.Args[1]
	os.Args = os.Args[1:] // Remove the command from os.Args for flag.Parse()

	switch command {
	case "server":
		socketPath := flag.String("socket", defaultSocketPath, "Unix domain socket path")
		flag.Parse()

		if err := server.Run(context.Background(), *socketPath); err != nil {
			log.Fatal(err)
		}

	case "client":
		socketPath := flag.String("socket", defaultSocketPath, "Unix domain socket path")
		workdir := flag.String("workdir", "", "Working directory (default: current directory)")
		flag.Parse()

		if flag.NArg() == 0 {
			log.Fatal("no command provided")
		}

		exitCode, err := client.Run(context.Background(), *socketPath, flag.Args(), *workdir)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(exitCode)

	default:
		log.Fatalf("Unknown command: %s. Please use 'server' or 'client'", command)
	}
}
