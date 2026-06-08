package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/supervisor"
	"github.com/sazardev/fugo/transport"
	"github.com/sazardev/fugo/ui"
)

func main() {
	addr := "127.0.0.1:9510"

	app := fugo.NewApp(fugo.AppOptions{Title: "Fugo Counter", Width: 800, Height: 600})

	server, _, err := transport.StartServer(addr, app)
	if err != nil {
		log.Fatalf("start server: %v", err)
	}
	defer server.GracefulStop()

	flutterBinary := findFlutterBinary()

	proc, err := supervisor.StartFlutter(context.Background(), addr, flutterBinary)
	if err != nil {
		log.Fatalf("start flutter: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("[fugo] shutting down")
		app.Shutdown()
		if err := proc.Shutdown(5 * time.Second); err != nil {
			log.Printf("[fugo] shutdown error: %v", err)
		}
		os.Exit(0)
	}()

	log.Println("[fugo] starting app")
	app.Run(buildCounterUI)
}

func buildCounterUI(ctx *fugo.Context) ui.Widget {
	counter := 0
	counterText := ui.NewText("0")

	incrementBtn := ui.NewButton("Increment").OnClick(func(e ui.Event) {
		counter++
		counterText.SetText(strconv.Itoa(counter))
		ctx.Update()
	})

	return ui.NewContainer(
		ui.NewCenter(
			ui.NewColumn(
				counterText,
				incrementBtn,
			),
		),
	)
}

func findFlutterBinary() string {
	candidates := []string{
		"flutter_client/build/linux/x64/debug/bundle/fugo_flutter_client",
		"flutter_client/build/linux/x64/release/bundle/fugo_flutter_client",
		os.Getenv("FUGO_FLUTTER_BINARY"),
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	log.Fatal("flutter binary not found. Run: cd flutter_client && flutter build linux")

	return ""
}
