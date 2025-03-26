package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"

	"connectrpc.com/connect"
	hashiruv1 "github.com/ryotarai/hashiru/gen/hashiru/v1"
	"github.com/ryotarai/hashiru/gen/hashiru/v1/hashiruv1connect"
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
		return 1, err
	}

	for stream.Receive() {
		res := stream.Msg()
		switch res.Result.(type) {
		case *hashiruv1.RunCommandResponse_StdoutBase64:
			stdout, err := base64.StdEncoding.DecodeString(res.GetStdoutBase64())
			if err != nil {
				return 1, err
			}
			fmt.Fprint(os.Stdout, string(stdout))
		case *hashiruv1.RunCommandResponse_StderrBase64:
			stderr, err := base64.StdEncoding.DecodeString(res.GetStderrBase64())
			if err != nil {
				return 1, err
			}
			fmt.Fprint(os.Stderr, string(stderr))
		case *hashiruv1.RunCommandResponse_ExitCode:
			return int(res.GetExitCode()), nil
		}
	}

	if err := stream.Err(); err != nil {
		return 1, err
	}

	return 0, nil
}
