package fugo

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/supervisor"
	"github.com/sazardev/fugo/transport"
	"github.com/sazardev/fugo/ui"
)

const schedulerInterval = 16 * time.Millisecond

type AppOptions struct {
	Title  string
	Width  int
	Height int
}

type App struct {
	uiRoot     ui.Widget
	ctx        *Context
	handlers   map[uint32]ui.Widget
	reconciler *engine.Reconciler
	scheduler  *engine.Scheduler
	oldTree    *fugov1.WidgetTree
	done       chan struct{}
	opts       AppOptions
}

type Context struct {
	app    *App
	router *ui.Router
}

func (c *Context) NavigateTo(route string) {
	if c.router != nil && c.router.NavigateTo(route) {
		c.Update()
	}
}

func (c *Context) GoBack() {
	if c.router != nil && c.router.GoBack() {
		c.Update()
	}
}

func (c *Context) Update() {
	c.app.scheduler.Enqueue()
}

func NewApp(opts AppOptions) *App {
	return &App{
		opts:      opts,
		handlers:  make(map[uint32]ui.Widget),
		done:      make(chan struct{}),
		scheduler: engine.NewScheduler(schedulerInterval),
	}
}

func (a *App) SetReconciler(stream engine.RenderStream) {
	if a.reconciler == nil {
		a.reconciler = engine.NewReconciler()
	}

	a.reconciler.SetStream(stream)
}

func (a *App) Run(buildUI func(ctx *Context) ui.Widget) {
	a.ctx = &Context{app: a}

	a.uiRoot = buildUI(a.ctx)

	if router, ok := a.uiRoot.(*ui.Router); ok {
		a.ctx.router = router
	}

	tree, widgetMap := ui.BuildTree(a.uiRoot)
	a.collectHandlers(widgetMap)
	a.oldTree = tree

	if a.reconciler == nil {
		a.reconciler = engine.NewReconciler()
	}

	a.reconciler.SendFullTree(tree)
	log.Println("[fugo] initial tree sent")

	a.scheduler.SetFlush(a.flush)
	a.scheduler.Start()

	<-a.done
	a.scheduler.Stop()
}

func (a *App) flush() {
	tree, widgetMap := ui.BuildTreeWithMerge(a.uiRoot, a.handlers)
	a.collectHandlers(widgetMap)

	patches := engine.Diff(a.oldTree, tree)
	if len(patches) > 0 {
		log.Printf("[fugo] flush: %d patches, %d nodes, %d handlers", len(patches), len(tree.Nodes), len(a.handlers))
		a.reconciler.SendPatches(patches)
	}

	a.oldTree = tree
}

func (a *App) Shutdown() {
	close(a.done)
}

func (a *App) HandleEvent(ev *fugov1.ClientEvent) {
	nodeID := parseNodeID(ev.GetNodeId())

	w, ok := a.handlers[nodeID]
	if !ok {
		log.Printf("[fugo] event: node %d not in handlers (%d registered, type=%s)", nodeID, len(a.handlers), ev.GetEventType())

		return
	}

	if !w.HasHandler() {
		return
	}

	log.Printf("[fugo] event: node=%d type=%s", nodeID, ev.GetEventType())
	w.Handle(ui.Event{
		NodeID:    ev.GetNodeId(),
		EventType: ev.GetEventType(),
		Data:      ev.GetEventData(),
	})
}

func (a *App) collectHandlers(m map[uint32]ui.Widget) {
	for id, w := range m {
		if w.HasHandler() {
			a.handlers[id] = w
		}
	}
}

func parseNodeID(s string) uint32 {
	var id uint32
	if _, err := fmt.Sscanf(s, "%d", &id); err != nil {
		return 0
	}

	return id
}

func RunStandalone(opts AppOptions, buildUI func(ctx *Context) ui.Widget) {
	app := NewApp(opts)

	addr := os.Getenv("FUGO_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9510"
	}

	server, _, err := transport.StartServer(addr, app)
	if err != nil {
		log.Fatalf("start server: %v", err)
	}

	flutterBinary := findFlutterBinary()
	if flutterBinary == "" {
		server.GracefulStop()
		log.Fatal("Flutter binary not found. Set FUGO_FLUTTER_BINARY env or build Flutter client.")
	}

	proc, err := supervisor.StartFlutter(context.Background(), addr, flutterBinary)
	if err != nil {
		server.GracefulStop()
		log.Fatalf("start flutter: %v", err)
	}

	go func() {
		<-proc.Exited()
		log.Println("[fugo] flutter window closed")
		app.Shutdown()
		server.GracefulStop()
		os.Exit(0)
	}()

	log.Println("[fugo] starting app")
	app.Run(buildUI)
}

func findFlutterBinary() string {
	if path := os.Getenv("FUGO_FLUTTER_BINARY"); path != "" {
		return path
	}

	// Search from CWD upward for fugo repo, then check flutter_client/
	dir, _ := os.Getwd()
	if fugoRoot := searchUpForFugo(dir); fugoRoot != "" {
		if fb := checkFugoFlutterDir(fugoRoot); fb != "" {
			return fb
		}
	}

	// Search from executable upward
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if fugoRoot := searchUpForFugo(exeDir); fugoRoot != "" {
			if fb := checkFugoFlutterDir(fugoRoot); fb != "" {
				return fb
			}
		}
	}

	// Common dev paths
	candidates := []string{
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "work", "fugo"),
		filepath.Join(os.Getenv("HOME"), "fugo"),
	}
	for _, path := range candidates {
		if fb := checkFugoFlutterDir(path); fb != "" {
			return fb
		}
	}

	return ""
}

func searchUpForFugo(start string) string {
	dir := start
	for {
		goMod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goMod); err == nil && strings.Contains(string(data), "github.com/sazardev/fugo") {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func checkFugoFlutterDir(fugoRoot string) string {
	candidates := []string{
		filepath.Join(fugoRoot, "flutter_client", "build", "windows", "x64", "runner", "Release", "fugo_flutter_client.exe"),
		filepath.Join(fugoRoot, "flutter_client", "build", "linux", "x64", "debug", "bundle", "fugo_flutter_client"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
