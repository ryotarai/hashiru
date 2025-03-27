package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	hashiruv1 "github.com/ryotarai/hashiru/gen/hashiru/v1"
)

type outputDest string

const (
	outputDestStdout outputDest = "stdout"
	outputDestStderr outputDest = "stderr"
)

type commandStreamWriter struct {
	stream *connect.ServerStream[hashiruv1.RunCommandResponse]
	dest   outputDest
}

func (w *commandStreamWriter) Write(p []byte) (int, error) {
	resp := &hashiruv1.RunCommandResponse{}
	switch w.dest {
	case outputDestStdout:
		resp.Result = &hashiruv1.RunCommandResponse_Stdout{
			Stdout: p,
		}
	case outputDestStderr:
		resp.Result = &hashiruv1.RunCommandResponse_Stderr{
			Stderr: p,
		}
	default:
		return 0, fmt.Errorf("invalid output destination: %s", w.dest)
	}
	err := w.stream.Send(resp)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type CommandServer struct{}

func (s *CommandServer) RunCommand(
	ctx context.Context,
	req *connect.Request[hashiruv1.RunCommandRequest],
	stream *connect.ServerStream[hashiruv1.RunCommandResponse],
) error {
	requestID := uuid.New().String()
	logger := slog.With("request_id", requestID)

	cmd := exec.CommandContext(ctx, req.Msg.GetName(), req.Msg.GetArgs()...)
	cmd.Env = req.Msg.GetEnv()
	cmd.Stdout = &commandStreamWriter{stream: stream, dest: outputDestStdout}
	cmd.Stderr = &commandStreamWriter{stream: stream, dest: outputDestStderr}
	cmd.Dir = req.Msg.GetDir()
	logger.Info("Running command", "name", req.Msg.GetName(), "args", req.Msg.GetArgs(), "dir", req.Msg.GetDir())
	err := cmd.Run()
	logger.Info("Command finished", "error", err)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			stream.Send(&hashiruv1.RunCommandResponse{
				Result: &hashiruv1.RunCommandResponse_ExitCode{
					ExitCode: int64(exitErr.ExitCode()),
				},
			})
		} else {
			return err
		}
	} else {
		stream.Send(&hashiruv1.RunCommandResponse{
			Result: &hashiruv1.RunCommandResponse_ExitCode{
				ExitCode: 0,
			},
		})
	}
	return nil
}
