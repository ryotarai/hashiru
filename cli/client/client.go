package client

import (
	"context"
	"encoding/base64"
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
func Run(ctx context.Context, socketPath string, args []string) (int, error) {
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

		dir, err := os.Getwd()
		if err != nil {
		return 1, err
	}

	req := &hashiruv1.RunCommandRequest{
		Name: args[0],
		Args: args[1:],
		Env:  os.Environ(),
		Dir:  dir,
	}

	stream, err := client.RunCommand(ctx, connect.NewRequest(req))
	if err != nil {
		return exitCodeForError, err
	}

	for stream.Receive() {
		res := stream.Msg()
		var base64Body string
		var output io.Writer
		switch res.Result.(type) {
		case *hashiruv1.RunCommandResponse_StdoutBase64:
			base64Body = res.GetStdoutBase64()
			output = os.Stdout
		case *hashiruv1.RunCommandResponse_StderrBase64:
			base64Body = res.GetStderrBase64()
			output = os.Stderr
		case *hashiruv1.RunCommandResponse_ExitCode:
			return int(res.GetExitCode()), nil
		default:
			return exitCodeForError, fmt.Errorf("unknown response: %v", res)
		}

		if output != nil {
			decoded, err := base64.StdEncoding.DecodeString(base64Body)
			if err != nil {
				return exitCodeForError, err
			}
			if _, err := output.Write(decoded); err != nil {
				return exitCodeForError, err
			}
		}
	}

	if err := stream.Err(); err != nil {
		return exitCodeForError, err
	}

	return 0, nil
}
