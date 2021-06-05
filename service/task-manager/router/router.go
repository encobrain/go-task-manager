package router

import (
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/lib/storage"
	"sync"
	"time"
)

func New(ctx context.Context) *Router {
	r := &Router{
		ctx:  ctx,
		subs: map[storage.Queue]*subs{},
	}

	return r
}

type routeState int

type subs struct {
	mu      sync.Mutex
	ctx     context.Context
	process chan storage.Task
	chans   map[string]chan storage.Task
	routed  map[storage.Task]bool
}

func (s *subs) getChannel(parrentUUID string) (ch chan storage.Task) {
	ch = s.chans[parrentUUID]

	if ch == nil {
		ch = make(chan storage.Task)
		s.chans[parrentUUID] = ch
	}

	return
}

// false - already routed
func (s *subs) setRouted(task storage.Task) (ok bool) {
	if s.routed[task] {
		return
	}

	s.routed[task] = true

	s.ctx.Child("canceled", func(ctx context.Context) {
		select {
		case <-ctx.Done():
		case <-task.Canceled():
			s.mu.Lock()
			defer s.mu.Unlock()

			delete(s.routed, task)
		}
	}).Go()

	return true
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
		subCtx := r.ctx.Child("subscribe", func(ctx context.Context) {
			tick := time.Tick(time.Second * 10)

			for {
				func() {
					tasks := queue.TasksGet()

					s.mu.Lock()
					defer s.mu.Unlock()

					for _, t := range tasks {
						if s.setRouted(t) {
							r.Route(queue, t)
						}
					}
				}()

				select {
				case <-ctx.Done():
					return
				case <-tick:
				}
			}
		})

		s = &subs{
			ctx:     subCtx,
			process: make(chan storage.Task),
			chans:   map[string]chan storage.Task{},
			routed:  map[storage.Task]bool{},
		}

		r.subs[queue] = s

		subCtx.Go()
	}

	return s.getChannel(parentUUID)
}

func (r *Router) Route(queue storage.Queue, task storage.Task) {
	canceled := task.Canceled()

	select {
	case <-canceled:
		return
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	s := r.subs[queue]

	if s == nil {
		return
	}

	allCh := s.getChannel("")
	parCh := s.getChannel(task.ParentUUID())

	r.ctx.Child("route", func(ctx context.Context) {
		s.mu.Lock()
		s.setRouted(task)
		s.mu.Unlock()

		defer func() { recover() }()

		select {
		case <-ctx.Done():
		case <-canceled:
		case allCh <- task:
		case parCh <- task:
		}
	}).Go()
}
