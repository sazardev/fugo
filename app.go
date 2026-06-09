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
	"sync"
	"time"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/fg"
	"github.com/sazardev/fugo/flog"
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
	// handlersMu guards handlers: the scheduler goroutine writes it from
	// flush→collectHandlers while the transport goroutine reads it in
	// HandleEvent.
	handlersMu sync.RWMutex

	// hostReqs correlates outstanding host-service requests (clipboard reads,
	// file dialogs) to the callback awaiting their reply. hostSeq mints the ids.
	hostMu   sync.Mutex
	hostSeq  uint64
	hostReqs map[uint64]func([]byte)
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

// UpdateNow is Update for latency-sensitive changes: it wakes the render loop
// immediately instead of waiting for the next frame, so the update reaches the
// client without up to a frame of delay. Prefer Update for ordinary mutations;
// reach for UpdateNow when the change must feel instant (e.g. echoing a
// keystroke). Like Update, repeated calls within the same instant coalesce.
func (c *Context) UpdateNow() {
	c.app.scheduler.EnqueueNow()
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
		hostReqs:  make(map[uint64]func([]byte)),
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
	flog.Infof("initial tree sent")

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
		flog.Debugf("flush: %d patches, %d nodes, %d handlers", len(patches), len(tree.GetNodes()), len(a.handlers))
		a.reconciler.SendPatches(patches)
	}

	a.oldTree = tree
}

// Shutdown stops the render loop and unblocks Run.
func (a *App) Shutdown() {
	close(a.done)
}

// hostEventType is the ClientEvent.event_type the client uses to reply to a
// HostCommand; the node_id then carries the request id rather than a widget id.
const hostEventType = "host"

// HandleEvent routes a client event to the handler of the widget whose node id
// matches. It implements the transport's app handler.
func (a *App) HandleEvent(ev *fugov1.ClientEvent) {
	if ev.GetEventType() == hostEventType {
		a.dispatchHostReply(ev)

		return
	}

	nodeID := parseNodeID(ev.GetNodeId())

	a.handlersMu.RLock()
	w, ok := a.handlers[nodeID]
	registered := len(a.handlers)
	a.handlersMu.RUnlock()

	if !ok {
		flog.Debugf("event: node %d not in handlers (%d registered, type=%s)", nodeID, registered, ev.GetEventType())

		return
	}

	if !w.HasHandler() {
		return
	}

	flog.Debugf("event: node=%d type=%s", nodeID, ev.GetEventType())
	w.Handle(fg.Event{
		NodeID:    ev.GetNodeId(),
		EventType: ev.GetEventType(),
		Data:      ev.GetEventData(),
	})
}

// sendHost issues a host-service command to the client. When cb is non-nil it
// is registered under a fresh request id and invoked with the reply bytes once
// the client answers; a nil cb means fire-and-forget (e.g. a clipboard write).
// The callback runs on the transport goroutine, like a widget event handler, so
// it may mutate widgets and call Context.Update.
func (a *App) sendHost(cmd *fugov1.HostCommand, cb func([]byte)) {
	if a.reconciler == nil {
		flog.Errorf("host command dropped: no client connected")

		return
	}

	if cb != nil {
		a.hostMu.Lock()
		a.hostSeq++
		id := a.hostSeq
		a.hostReqs[id] = cb
		a.hostMu.Unlock()

		cmd.RequestId = id
	}

	a.reconciler.SendHostCommand(cmd)
}

func (a *App) dispatchHostReply(ev *fugov1.ClientEvent) {
	id, err := strconv.ParseUint(ev.GetNodeId(), 10, 64)
	if err != nil {
		flog.Errorf("host reply with bad request id %q: %v", ev.GetNodeId(), err)

		return
	}

	a.hostMu.Lock()
	cb, ok := a.hostReqs[id]
	delete(a.hostReqs, id)
	a.hostMu.Unlock()

	if ok && cb != nil {
		cb(ev.GetEventData())
	}
}

func (a *App) collectHandlers(m map[uint32]fg.Widget) {
	a.handlersMu.Lock()
	defer a.handlersMu.Unlock()

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

	tuneRuntime()
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

	if os.Getenv("FUGO_NO_FLUTTER") == "1" {
		// Server-only mode: `fugo run --watch` owns the Flutter process across
		// reloads (the window stays open and reconnects), so here we just serve
		// and block until the watcher restarts us.
		flog.Infof("starting app (server-only; window managed by the watcher)")
		app.Run(buildUI)

		return
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
		flog.Infof("flutter window closed")
		app.Shutdown()
		server.GracefulStop()
		os.Exit(0)
	}()

	flog.Infof("starting app")
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
		flog.Errorf("could not generate auth token: %v", err)

		return
	}

	_ = os.Setenv("FUGO_TOKEN", hex.EncodeToString(buf))
	flog.Infof("per-run auth token enabled (FUGO_AUTH=1)")
}

// exportWindowEnv forwards the window options and the active theme to the
// Flutter client through the environment (FUGO_TITLE/WIDTH/HEIGHT and
// FUGO_THEME_SEED/BRIGHTNESS); the supervisor passes them along. The client
// builds its Material 3 ColorScheme from the seed + brightness, so call
// fg.UseTheme before RunStandalone to change it.
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

	theme := fg.CurrentTheme()
	_ = os.Setenv("FUGO_THEME_SEED", theme.Colors.Primary.String())
	_ = os.Setenv("FUGO_THEME_BRIGHTNESS", theme.Brightness())
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
