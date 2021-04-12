package sqlite

import (
	"fmt"
	"github.com/encobrain/go-task-manager/lib/storage"
	confStorage "github.com/encobrain/go-task-manager/model/config/lib/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"
)

func New(config *confStorage.SQLite) *Storage {
	s := &Storage{
		conf: config,
	}

	return s
}

type Storage struct {
	conf *confStorage.SQLite

	mu        sync.Mutex
	isStarted bool
	db        *gorm.DB

	cache struct {
		queueByName sync.Map // map[name]*queue
		queueById   sync.Map // map[uint64]*queue
	}
}

func (s *Storage) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStarted {
		return
	}

	log.Printf("SQLiteStorage: starting...\n")

	db, err := gorm.Open(sqlite.Open(s.conf.Db.Path))

	if err != nil {
		panic(fmt.Errorf("open db `%s` fail. %s", s.conf.Db.Path, err))
	}

	err = db.AutoMigrate(&dbQueue{}, &dbTask{})

	if err != nil {
		panic(fmt.Errorf("auto migrate fail. %s", err))
	}

	s.db = db
	s.isStarted = true

	log.Printf("SQLiteStorage: started\n")
}

func (s *Storage) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isStarted {
		return
	}

	log.Printf("SQLiteStorage: stopping...\n")

	s.isStarted = false

	log.Printf("SQLiteStorage: stopped\n")
}

func (s *Storage) QueueGetOrCreate(name string) *storage.QueueInfo {
	iq, _ := s.cache.queueByName.Load(name)

	for iq == nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.isStarted {
			panic(fmt.Errorf("storage stopped"))
		}

		iq, _ = s.cache.queueByName.Load(name)

		if iq != nil {
			break
		}

		dbq := &dbQueue{Name: name}
		err := s.db.Where(dbq).Take(dbq).Error

		if err != nil {
			if err != gorm.ErrRecordNotFound {
				panic(fmt.Errorf("get queue `%s` from storage fail. %s", name, err))
			}

			err = s.db.Create(dbq).Error

			if err != nil {
				panic(fmt.Errorf("create queue `%s` fail. %s", name, err))
			}
		}

		q := &queue{dbQueue: dbq, db: s.db}
		q.task.all = true
		q.start()

		s.cache.queueByName.Store(name, q)
		s.cache.queueById.Store(uint64(dbq.ID), q)

		iq = q
		break
	}

	q := iq.(*queue)

	return &storage.QueueInfo{
		Id:   uint64(q.ID),
		Name: q.Name,
	}
}

func (s *Storage) QueueGet(id uint64) *storage.QueueInfo {
	iq, _ := s.cache.queueById.Load(id)

	for iq == nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.isStarted {
			panic(fmt.Errorf("storage stopped"))
		}

		iq, _ = s.cache.queueById.Load(id)

		if iq != nil {
			break
		}

		dbq := &dbQueue{}
		dbq.ID = uint(id)
		err := s.db.Where(dbq).Take(dbq).Error

		if err != nil {
			if err != gorm.ErrRecordNotFound {
				panic(fmt.Errorf("get queue `%d` from storage fail. %s", id, err))
			}

			return nil
		}

		q := &queue{dbQueue: dbq, db: s.db}
		q.start()

		s.cache.queueByName.Store(q.Name, q)
		s.cache.queueById.Store(uint64(dbq.ID), q)

		iq = q
		break
	}

	q := iq.(*queue)

	return &storage.QueueInfo{
		Id:   uint64(q.ID),
		Name: q.Name,
	}
}

func (s *Storage) TaskNew(queueId uint64, parentUUID string, status string, content []byte) *storage.TaskInfo {
	iq, _ := s.cache.queueById.Load(queueId)

	if iq == nil {
		return nil
	}

	return iq.(*queue).taskNew(parentUUID, status, content)
}

func (s *Storage) TaskGet(queueId uint64, uuid string) *storage.TaskInfo {
	iq, _ := s.cache.queueById.Load(queueId)

	if iq == nil {
		return nil
	}

	return iq.(*queue).taskGetInfo(uuid)
}

func (s *Storage) TasksGet(queueId uint64) []*storage.TaskInfo {
	iq, _ := s.cache.queueById.Load(queueId)

	if iq == nil {
		return nil
	}

	return iq.(*queue).tasksInfoGet()
}

func (s *Storage) TaskStatusSet(queueId uint64, uuid string, status string, content []byte) (ok bool) {
	iq, _ := s.cache.queueById.Load(queueId)

	if iq == nil {
		return
	}

	return iq.(*queue).taskStatusSet(uuid, status, content)
}

func (s *Storage) TaskRemove(queueId uint64, uuid string) (ok bool) {
	iq, _ := s.cache.queueById.Load(queueId)

	if iq == nil {
		return
	}

	return iq.(*queue).taskRemove(uuid)
}
