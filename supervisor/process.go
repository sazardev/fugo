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
	cmd    *exec.Cmd
	addr   string
	exited chan struct{}
}

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

	log.Printf("[fugo] flutter client started (pid=%d)", cmd.Process.Pid)

	fp := &FlutterProcess{
		cmd:    cmd,
		addr:   addr,
		exited: make(chan struct{}),
	}

	go func() {
		_ = cmd.Wait()
		close(fp.exited)
		log.Println("[fugo] flutter client exited")
	}()

	return fp, nil
}

func (p *FlutterProcess) Exited() <-chan struct{} {
	return p.exited
}

func (p *FlutterProcess) WaitForSignal(timeout time.Duration) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("[fugo] received signal %v, shutting down", sig)

	return p.Shutdown(timeout)
}

func (p *FlutterProcess) Shutdown(timeout time.Duration) error {
	log.Println("[fugo] shutting down flutter client")

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("[fugo] signal error: %v", err)
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
		if err := p.cmd.Process.Kill(); err != nil {
			log.Printf("[fugo] force kill error: %v", err)
		}
		<-done
	}

	return nil
}
