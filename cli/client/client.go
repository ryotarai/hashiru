package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"connectrpc.com/connect"
	hashiruv1 "github.com/ryotarai/hashiru/gen/hashiru/v1"
	"github.com/ryotarai/hashiru/gen/hashiru/v1/hashiruv1connect"
)

const (
	exitCodeForError = 125
)

// Run executes a command through the server and returns the exit code.
func Run(ctx context.Context, socketPath string, args []string, workdir string) (int, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	client := hashiruv1connect.NewCommandServiceClient(
		httpClient,
		"http://unix",
	)

	if workdir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return exitCodeForError, err
		}
		workdir = dir
	}

	req := &hashiruv1.RunCommandRequest{
		Name: args[0],
		Args: args[1:],
		Env:  os.Environ(),
		Dir:  workdir,
	}

	stream, err := client.RunCommand(ctx, connect.NewRequest(req))
	if err != nil {
		return exitCodeForError, err
	}

	for stream.Receive() {
		res := stream.Msg()
		var output io.Writer
		var body []byte
		switch res.Result.(type) {
		case *hashiruv1.RunCommandResponse_Stdout:
			output = os.Stdout
			body = res.GetStdout()
		case *hashiruv1.RunCommandResponse_Stderr:
			output = os.Stderr
			body = res.GetStderr()
		case *hashiruv1.RunCommandResponse_ExitCode:
			return int(res.GetExitCode()), nil
		default:
			return exitCodeForError, fmt.Errorf("unknown response: %v", res)
		}

		if output != nil {
			if _, err := output.Write(body); err != nil {
				return exitCodeForError, err
			}
		}
	}

	if err := stream.Err(); err != nil {
		return exitCodeForError, err
	}

	return 0, nil
}
