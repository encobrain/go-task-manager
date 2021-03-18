package queue

import "github.com/encobrain/go-task-manager/lib/storage"

type Manager interface {
	// GetOrCreate gets queue by name or creates it if not exists
	GetOrCreate(name string) storage.Queue
	// Get gets queue by id.
	// If nil - queue not exists
	Get(id uint64) storage.Queue
}
