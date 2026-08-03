package fg

import (
	"sync"
	"testing"
)

func TestStore_GetSet(t *testing.T) {
	s := NewStore(0)
	if got := s.Get(); got != 0 {
		t.Fatalf("Get() = %d, want 0", got)
	}

	s.Set(5)
	if got := s.Get(); got != 5 {
		t.Fatalf("Get() = %d, want 5", got)
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore(struct{ Count int }{})

	s.Update(func(v *struct{ Count int }) { v.Count++ })
	s.Update(func(v *struct{ Count int }) { v.Count++ })

	if got := s.Get().Count; got != 2 {
		t.Fatalf("Count = %d, want 2", got)
	}
}

func TestStore_Subscribe(t *testing.T) {
	s := NewStore(0)

	var got []int

	s.Subscribe(func(v int) { got = append(got, v) })
	s.Set(1)
	s.Update(func(v *int) { *v = 2 })

	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("subscriber saw %v, want [1 2]", got)
	}
}

func TestStore_ConcurrentUpdate(t *testing.T) {
	s := NewStore(0)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Update(func(v *int) { *v++ })
		}()
	}
	wg.Wait()

	if got := s.Get(); got != 100 {
		t.Fatalf("Get() = %d, want 100", got)
	}
}
