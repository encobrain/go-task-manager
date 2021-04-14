package queue

import (
	"github.com/encobrain/go-task-manager/lib/storage"
	"sync"
)

type Manager interface {
	// GetOrCreate gets queue by name or creates it if not exists
	GetOrCreate(name string) storage.Queue
	// Get gets queue by id.
	// If nil - queue not exists
	Get(id uint64) storage.Queue
}

func NewManager(stor storage.Storage) *manager {
	m := &manager{
		storage: stor,
	}

	m.queue.byName = map[string]storage.Queue{}
	m.queue.byId = map[uint64]storage.Queue{}

	return m
}

type manager struct {
	storage storage.Storage

	mu    sync.Mutex
	queue struct {
		byName map[string]storage.Queue
		byId   map[uint64]storage.Queue
	}
}

func (m *manager) GetOrCreate(name string) storage.Queue {
	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.queue.byName[name]

	if q == nil {
		qi := m.storage.QueueGetOrCreate(name)

		q = storage.NewQueue(m.storage, qi)
		m.queue.byName[name] = q
		m.queue.byId[qi.Id] = q
	}

	return q
}

func (m *manager) Get(id uint64) storage.Queue {
	m.mu.Lock()
	defer m.mu.Unlock()

	q := m.queue.byId[id]

	if q == nil {
		qi := m.storage.QueueGet(id)

		if qi == nil {
			return nil
		}

		q = storage.NewQueue(m.storage, qi)
		m.queue.byName[qi.Name] = q
		m.queue.byId[qi.Id] = q
	}

	return q
}
