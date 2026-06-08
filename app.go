package fugo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/fg"
	"github.com/sazardev/fugo/supervisor"
	"github.com/sazardev/fugo/transport"
)

const schedulerInterval = 16 * time.Millisecond

// AppOptions configures the window the Flutter client opens.
type AppOptions struct {
	Title  string
	Width  int
	Height int
}

// App owns the retained widget tree, the event-handler registry, and the
// render loop (scheduler → diff → reconciler) that drives the client.
type App struct {
	uiRoot     fg.Widget
	ctx        *Context
	handlers   map[uint32]fg.Widget
	reconciler *engine.Reconciler
	scheduler  *engine.Scheduler
	oldTree    *fugov1.WidgetTree
	done       chan struct{}
	opts       AppOptions
}

// Context is passed to buildUI and to event handlers. It exposes navigation
// and Update, the call that schedules a re-render after mutating widgets.
type Context struct {
	app    *App
	router *fg.RouterWidget
}

// NavigateTo asks the active router to switch to route and re-renders.
func (c *Context) NavigateTo(route string) {
	if c.router != nil && c.router.NavigateTo(route) {
		c.Update()
	}
}

// GoBack pops the router history and re-renders if a previous route exists.
func (c *Context) GoBack() {
	if c.router != nil && c.router.GoBack() {
		c.Update()
	}
}

// Update marks the UI dirty so the scheduler diffs and flushes on the next
// frame. Call it after mutating widgets in an event handler.
func (c *Context) Update() {
	c.app.scheduler.Enqueue()
}

// Param returns the value captured for a :param in the active route (e.g. "id"
// for a route registered as "/user/:id"), or "" if there is no router or no
// such parameter.
func (c *Context) Param(name string) string {
	if c.router != nil {
		return c.router.Param(name)
	}

	return ""
}

// Window returns a controller for the client's OS window.
func (c *Context) Window() *WindowController {
	return &WindowController{app: c.app}
}

// WindowController drives the client's OS window at runtime. Obtain one via
// Context.Window. Commands are streamed to the Flutter client, which applies
// them through its window manager.
type WindowController struct {
	app *App
}

func (w *WindowController) send(cmd *fugov1.WindowCommand) {
	if w.app.reconciler != nil {
		w.app.reconciler.SendWindowCommand(cmd)
	}
}

// SetTitle changes the window title.
func (w *WindowController) SetTitle(title string) {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_SET_TITLE, Title: title})
}

// SetSize resizes the window to width x height logical pixels.
func (w *WindowController) SetSize(width, height float64) {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_SET_SIZE, Width: width, Height: height})
}

// Minimize minimizes the window to the taskbar.
func (w *WindowController) Minimize() {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_MINIMIZE})
}

// Maximize maximizes the window.
func (w *WindowController) Maximize() {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_MAXIMIZE})
}

// Center centers the window on the current screen.
func (w *WindowController) Center() {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_CENTER})
}

// SetFullScreen enables or disables fullscreen mode.
func (w *WindowController) SetFullScreen(on bool) {
	w.send(&fugov1.WindowCommand{Op: fugov1.WindowOp_WINDOW_FULLSCREEN, Flag: on})
}

// Component renders a UI from a value that can carry state, as an alternative
// to a buildUI closure. Implement Render and pass the component to RunComponent
// (or app.Run(c.Render)); event handlers mutate the component's fields and the
// widgets it built, then call Context.Update — the same retained-tree model.
type Component interface {
	Render(ctx *Context) fg.Widget
}

// NewApp creates an App with the given options and a 60fps scheduler. Use
// RunStandalone for the common case of also starting the server and client.
func NewApp(opts AppOptions) *App {
	return &App{
		opts:      opts,
		handlers:  make(map[uint32]fg.Widget),
		done:      make(chan struct{}),
		scheduler: engine.NewScheduler(schedulerInterval),
	}
}

// SetReconciler binds the app's reconciler to a render stream (called by the
// transport when a client connects).
func (a *App) SetReconciler(stream engine.RenderStream) {
	if a.reconciler == nil {
		a.reconciler = engine.NewReconciler()
	}

	a.reconciler.SetStream(stream)
}

// Run builds the retained widget tree once via buildUI, sends the initial
// tree, then starts the scheduler and blocks until Shutdown.
func (a *App) Run(buildUI func(ctx *Context) fg.Widget) {
	a.ctx = &Context{app: a}

	a.uiRoot = buildUI(a.ctx)

	if router, ok := a.uiRoot.(*fg.RouterWidget); ok {
		a.ctx.router = router
	}

	tree, widgetMap := fg.BuildTree(a.uiRoot)
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
	tree, widgetMap := fg.BuildTreeWithMerge(a.uiRoot, a.handlers)
	a.collectHandlers(widgetMap)

	patches := engine.Diff(a.oldTree, tree)
	if len(patches) > 0 {
		log.Printf("[fugo] flush: %d patches, %d nodes, %d handlers", len(patches), len(tree.GetNodes()), len(a.handlers))
		a.reconciler.SendPatches(patches)
	}

	a.oldTree = tree
}

// Shutdown stops the render loop and unblocks Run.
func (a *App) Shutdown() {
	close(a.done)
}

// HandleEvent routes a client event to the handler of the widget whose node id
// matches. It implements the transport's app handler.
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
	w.Handle(fg.Event{
		NodeID:    ev.GetNodeId(),
		EventType: ev.GetEventType(),
		Data:      ev.GetEventData(),
	})
}

func (a *App) collectHandlers(m map[uint32]fg.Widget) {
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

// RunStandalone is the one-call entry point: it starts the gRPC server, spawns
// the Flutter client, builds the UI, and runs until the window closes.
func RunStandalone(opts AppOptions, buildUI func(ctx *Context) fg.Widget) {
	app := NewApp(opts)

	enableAuthToken()
	exportWindowEnv(opts)

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

// RunComponent is RunStandalone for a Component: it renders c.Render.
func RunComponent(opts AppOptions, c Component) {
	RunStandalone(opts, c.Render)
}

// enableAuthToken generates a per-run token when FUGO_AUTH=1 so the transport
// rejects any local process that does not present it. It is opt-in to avoid
// breaking a Flutter client that was built without token support.
func enableAuthToken() {
	if os.Getenv("FUGO_AUTH") != "1" || os.Getenv("FUGO_TOKEN") != "" {
		return
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("[fugo] could not generate auth token: %v", err)

		return
	}

	_ = os.Setenv("FUGO_TOKEN", hex.EncodeToString(buf))
	log.Println("[fugo] per-run auth token enabled (FUGO_AUTH=1)")
}

// exportWindowEnv forwards the window options to the Flutter client through the
// environment (FUGO_TITLE/WIDTH/HEIGHT); the supervisor passes them along.
func exportWindowEnv(opts AppOptions) {
	if opts.Title != "" {
		_ = os.Setenv("FUGO_TITLE", opts.Title)
	}

	if opts.Width > 0 {
		_ = os.Setenv("FUGO_WIDTH", strconv.Itoa(opts.Width))
	}

	if opts.Height > 0 {
		_ = os.Setenv("FUGO_HEIGHT", strconv.Itoa(opts.Height))
	}
}

func findFlutterBinary() string {
	if path := os.Getenv("FUGO_FLUTTER_BINARY"); path != "" {
		return path
	}

	// A packaged app (produced by `fugo build`) ships the Flutter client in a
	// flutter/ folder next to the executable. Prefer that — it makes the app
	// self-contained and independent of the fugo source tree.
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, p := range []string{
			filepath.Join(exeDir, "flutter", "fugo_flutter_client.exe"),
			filepath.Join(exeDir, "flutter", "fugo_flutter_client"),
		} {
			if _, statErr := os.Stat(p); statErr == nil {
				return p
			}
		}
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
		if data, err := os.ReadFile(goMod); err == nil && strings.Contains(string(data), "github.com/sazardev/fugo") { //nolint:gosec // path derived from os.Getwd walk, not external input
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
		if _, err := os.Stat(path); err == nil { //nolint:gosec // path built from fugoRoot via filepath.Join, not external input
			return path
		}
	}

	return ""
}
