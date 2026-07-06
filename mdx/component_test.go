package mdx

import (
	"sync"
	"testing"
)

// Run with -race: exercises Register (write) and Transpile (read) from
// multiple goroutines against the same ComponentRegistry.
func TestComponentRegistryConcurrentAccess(t *testing.T) {
	r := NewComponentRegistry()

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Register("Custom", func(attrs map[string]string, inner string) string {
				return "<div>" + inner + "</div>"
			})
		}(i)
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Transpile(`<DeepDive title="T">content</DeepDive>`)
		}()
	}

	wg.Wait()
}
