package fg

import "sync"

// Store is an opt-in, generic piece of shared state for apps whose state is
// touched by more than a couple of widgets — most Fugo apps don't need it:
// mutating a widget struct in a handler and calling ctx.Update() is enough.
// Reach for a Store when several widgets across the tree need to react to the
// same underlying data, when you want a single place to inspect "current
// state" for a test, or when handlers run concurrently (an OnChange closure
// on the transport goroutine, a Notifications/Clipboard callback also on the
// transport goroutine, the scheduler's flush goroutine) and need to touch the
// same data safely.
//
// A Store does not replace Context or the retained widget tree: subscribers
// are just closures that copy fields out of the new state into widget
// setters and call ctx.Update() themselves, exactly as any other handler
// would. Example:
//
//	type appState struct{ count int }
//
//	store := fg.NewStore(appState{})
//	countText := fg.Text("0")
//	store.Subscribe(func(s appState) {
//		countText.SetText(strconv.Itoa(s.count))
//		ctx.Update()
//	})
//
//	incBtn := fg.Button("+1").OnClick(func(fg.Event) {
//		store.Update(func(s *appState) { s.count++ })
//	})
type Store[S any] struct {
	mu    sync.Mutex
	state S
	subs  []func(S)
}

// NewStore creates a Store holding initial as its starting state.
func NewStore[S any](initial S) *Store[S] {
	return &Store[S]{state: initial}
}

// Get returns a copy of the current state. Safe to call from any goroutine.
func (s *Store[S]) Get() S {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state
}

// Set replaces the state wholesale and notifies every subscriber with the
// new value, in registration order. Safe to call from any goroutine.
func (s *Store[S]) Set(next S) {
	s.mu.Lock()
	s.state = next
	subs := make([]func(S), len(s.subs))
	copy(subs, s.subs)
	s.mu.Unlock()

	for _, fn := range subs {
		fn(next)
	}
}

// Update runs fn against a pointer to the current state (mutate it in
// place — a reducer-style ergonomic entry point) and then notifies every
// subscriber with the result. Safe to call from any goroutine.
func (s *Store[S]) Update(fn func(*S)) {
	s.mu.Lock()
	fn(&s.state)
	next := s.state
	subs := make([]func(S), len(s.subs))
	copy(subs, s.subs)
	s.mu.Unlock()

	for _, sub := range subs {
		sub(next)
	}
}

// Subscribe registers fn to be called with the new state on every Set/Update.
// fn is not called immediately with the current state — call Get yourself
// first if you need to initialize widgets before the first change.
func (s *Store[S]) Subscribe(fn func(S)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subs = append(s.subs, fn)
}
