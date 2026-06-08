package fugo

import (
	"fmt"
	"log"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/ui"
)

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
	oldTree    *fugov1.WidgetTree
	done       chan struct{}
	opts       AppOptions
}

type Context struct {
	app      *App
	updateCh chan struct{}
}

func NewApp(opts AppOptions) *App {
	return &App{
		opts:     opts,
		handlers: make(map[uint32]ui.Widget),
		done:     make(chan struct{}),
	}
}

func (a *App) SetReconciler(stream engine.RenderStream) {
	if a.reconciler == nil {
		a.reconciler = engine.NewReconciler()
	}

	a.reconciler.SetStream(stream)
}

func (a *App) Run(buildUI func(ctx *Context) ui.Widget) {
	a.ctx = &Context{app: a, updateCh: make(chan struct{}, 8)}

	a.uiRoot = buildUI(a.ctx)

	tree, widgetMap := ui.BuildTree(a.uiRoot)
	a.collectHandlers(widgetMap)
	a.oldTree = tree

	if a.reconciler == nil {
		a.reconciler = engine.NewReconciler()
	}

	a.reconciler.SendFullTree(tree)
	log.Println("[fugo] initial tree sent")

	for {
		select {
		case <-a.ctx.updateCh:
			tree, widgetMap = ui.BuildTreeWithMerge(a.uiRoot, a.handlers)
			a.collectHandlers(widgetMap)

			patches := engine.Diff(a.oldTree, tree)
			if len(patches) > 0 {
				a.reconciler.SendPatches(patches)
			}

			a.oldTree = tree
		case <-a.done:
			return
		}
	}
}

func (a *App) Shutdown() {
	close(a.done)
}

func (a *App) HandleEvent(ev *fugov1.ClientEvent) {
	nodeID := parseNodeID(ev.GetNodeId())

	w, ok := a.handlers[nodeID]
	if !ok {
		return
	}

	if !w.HasHandler() {
		return
	}

	w.Handle(ui.Event{
		NodeID:    ev.GetNodeId(),
		EventType: ev.GetEventType(),
		Data:      ev.GetEventData(),
	})
}

func (c *Context) Update() {
	select {
	case c.updateCh <- struct{}{}:
	default:
	}
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
