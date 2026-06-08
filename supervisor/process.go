package supervisor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type FlutterProcess struct {
	cmd  *exec.Cmd
	addr string
}

func StartFlutter(ctx context.Context, addr, flutterBinary string) (*FlutterProcess, error) {
	cmd := exec.CommandContext(ctx, flutterBinary)

	cmd.Env = append(
		os.Environ(),
		"FUGO_ADDR="+addr,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGTERM,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start flutter: %w", err)
	}

	log.Printf("[fugo] flutter client started (pid=%d)", cmd.Process.Pid)

	return &FlutterProcess{
		cmd:  cmd,
		addr: addr,
	}, nil
}

func (p *FlutterProcess) WaitForSignal(timeout time.Duration) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-sigCh
	log.Printf("[fugo] received signal %v, shutting down", sig)
	return p.Shutdown(timeout)
}

func (p *FlutterProcess) Shutdown(timeout time.Duration) error {
	log.Println("[fugo] shutting down flutter client")

	if err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM); err != nil {
		log.Printf("[fugo] kill error: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("[fugo] flutter exited with error: %v", err)
		}
	case <-time.After(timeout):
		log.Println("[fugo] flutter didn't exit, force killing")
		if err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Printf("[fugo] force kill error: %v", err)
		}
		<-done
	}

	return nil
}
