package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sazardev/fugo/flog"
)

// FlutterProcess represents the spawned Flutter render client subprocess and
// tracks its lifecycle, exposing channels and methods to wait for exit and to
// shut it down gracefully.
type FlutterProcess struct {
	cmd    *exec.Cmd
	addr   string
	exited chan struct{}
}

// StartFlutter launches the Flutter render client at flutterBinary, passing the
// gRPC address via the FUGO_ADDR environment variable and inheriting the
// parent's stdout/stderr. The process is tied to ctx; a goroutine monitors it
// and closes the Exited channel when it terminates.
func StartFlutter(ctx context.Context, addr, flutterBinary string) (*FlutterProcess, error) {
	cmd := exec.CommandContext(ctx, flutterBinary)

	cmd.Env = append(
		os.Environ(),
		"FUGO_ADDR="+addr,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start flutter: %w", err)
	}

	flog.Infof("flutter client started (pid=%d)", cmd.Process.Pid)

	fp := &FlutterProcess{
		cmd:    cmd,
		addr:   addr,
		exited: make(chan struct{}),
	}

	go func() {
		_ = cmd.Wait()
		close(fp.exited)
		flog.Infof("flutter client exited")
	}()

	return fp, nil
}

// Exited returns a channel that is closed when the Flutter subprocess exits.
func (p *FlutterProcess) Exited() <-chan struct{} {
	return p.exited
}

// WaitForSignal blocks until an interrupt or termination signal is received,
// then shuts the subprocess down, allowing up to timeout for a clean exit.
func (p *FlutterProcess) WaitForSignal(timeout time.Duration) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	flog.Infof("received signal %v, shutting down", sig)

	return p.Shutdown(timeout)
}

// Shutdown stops the Flutter subprocess by sending SIGTERM and waiting up to
// timeout for it to exit; if it does not exit in time it is force-killed.
func (p *FlutterProcess) Shutdown(timeout time.Duration) error {
	flog.Infof("shutting down flutter client")

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		flog.Errorf("signal error: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			flog.Errorf("flutter exited with error: %v", err)
		}
	case <-time.After(timeout):
		flog.Infof("flutter didn't exit, force killing")
		if err := p.cmd.Process.Kill(); err != nil {
			flog.Errorf("force kill error: %v", err)
		}
		<-done
	}

	return nil
}
