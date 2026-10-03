package generator

import (
	"sync"

	"github.com/yasunori0418/layat/internal/manifest"
)

// Fake is the test double of the contract: each operation returns what its stub func returns
// (zero values when the stub is nil) and is recorded in call order. It lives outside _test files
// so the cmd layer's tests can inject it too. Safe for concurrent use (apply --all builds in
// parallel).
type Fake struct {
	DiscoverFunc func(file string) error
	RootsFunc    func(name string) (manifest.Root, error)
	AllRootsFunc func() (map[string]manifest.Root, error)
	BuildFunc    func(name, pending string) (string, error)
	DryBuildFunc func(name string) (string, error)

	mu    sync.Mutex
	calls []Call
}

// Call is one recorded operation: its method name and string arguments.
type Call struct {
	Op   string
	Args []string
}

// Calls returns a copy of the recorded operations in call order.
func (f *Fake) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Call(nil), f.calls...)
}

func (f *Fake) record(op string, args ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Op: op, Args: args})
}

func (f *Fake) Discover(file string) error {
	f.record("Discover", file)
	if f.DiscoverFunc == nil {
		return nil
	}
	return f.DiscoverFunc(file)
}

func (f *Fake) Roots(name string) (manifest.Root, error) {
	f.record("Roots", name)
	if f.RootsFunc == nil {
		return manifest.Root{}, nil
	}
	return f.RootsFunc(name)
}

func (f *Fake) AllRoots() (map[string]manifest.Root, error) {
	f.record("AllRoots")
	if f.AllRootsFunc == nil {
		return nil, nil
	}
	return f.AllRootsFunc()
}

func (f *Fake) Build(name, pending string) (string, error) {
	f.record("Build", name, pending)
	if f.BuildFunc == nil {
		return "", nil
	}
	return f.BuildFunc(name, pending)
}

func (f *Fake) DryBuild(name string) (string, error) {
	f.record("DryBuild", name)
	if f.DryBuildFunc == nil {
		return "", nil
	}
	return f.DryBuildFunc(name)
}
