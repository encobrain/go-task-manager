package task_manager

import (
	"github.com/encobrain/go-task-manager/service"
	"log"
)

func (s *tmService) Status() (status service.Status) {
	s.status.lock.RLock()
	defer s.status.lock.RUnlock()

	status = s.status.v
	return
}

func (s *tmService) StatusWait(status service.Status) (waiter <-chan struct{}) {
	s.status.lock.Lock()
	defer s.status.lock.Unlock()

	ch := make(chan struct{})

	if s.status.v == status {
		close(ch)
		return ch
	}

	s.status.waiters[status] = append(s.status.waiters[status], ch)

	return ch
}

func (s *tmService) statusSet(new service.Status, prev ...service.Status) bool {
	s.status.lock.Lock()
	defer s.status.lock.Unlock()

	if s.status.v == new {
		return true
	}

	if len(prev) != 0 {
		ok := false
		for _, p := range prev {
			if s.status.v == p {
				ok = true
				break
			}
		}

		if !ok {
			return false
		}
	}

	s.status.v = new

	log.Printf("Status: %s", new)

	waiters := s.status.waiters[new]

	if waiters != nil {
		delete(s.status.waiters, new)

		for _, ch := range waiters {
			close(ch)
		}
	}

	return true
}
