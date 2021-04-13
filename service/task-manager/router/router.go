package router

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/storage"
	"sync"
)

func New(ctx context.Context) *Router {
	r := &Router{
		ctx:  ctx,
		subs: map[storage.Queue]*subs{},
	}

	return r
}

type subs struct {
	process chan storage.Task
	chans   map[string]chan storage.Task
}

func (s *subs) getChannel(parrentUUID string) (ch chan storage.Task) {
	ch = s.chans[parrentUUID]

	if ch == nil {
		ch = make(chan storage.Task)
		s.chans[parrentUUID] = ch
	}

	return
}

type Router struct {
	ctx  context.Context
	mu   sync.Mutex
	subs map[storage.Queue]*subs
}

func (r *Router) Subscribe(queue storage.Queue, parentUUID string) (tasks <-chan storage.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := r.subs[queue]

	if s == nil {
		s = &subs{
			process: make(chan storage.Task),
			chans:   map[string]chan storage.Task{},
		}

		r.subs[queue] = s

		tasks := queue.TasksGet()

		r.ctx.Child("subscribe", func(ctx context.Context) {
			for _, t := range tasks {
				r.Route(queue, t)
			}
		}).Go()
	}

	return s.getChannel(parentUUID)
}

func (r *Router) Route(queue storage.Queue, task storage.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := r.subs[queue]

	if s == nil {
		return
	}

	allCh := s.getChannel("")
	parCh := s.getChannel(task.ParentUUID())

	r.ctx.Child("route", func(ctx context.Context) {
		defer func() { recover() }()

		select {
		case <-ctx.Done():
		case <-task.Canceled():
		case allCh <- task:
		case parCh <- task:
		}
	}).Go()
}
