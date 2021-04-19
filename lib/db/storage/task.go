package storage

import "sync"

type task struct {
	mu sync.Mutex

	*dbTask

	updated bool
}

func (t *task) statusSet(status string, content []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = status
	t.Content = content
	t.updated = true
}
