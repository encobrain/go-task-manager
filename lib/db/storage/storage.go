package storage

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"sync"
)

type QueueInfo struct {
	Id   uint64
	Name string
}

type TaskInfo struct {
	QueueId    uint64
	UUID       string
	ParentUUID string
	Status     string
	Content    []byte
}

type Storage interface {
	// QueueGetOrCreate gets queue by name or creates it if not exists
	QueueGetOrCreate(name string) *QueueInfo
	// QueueGet gets queue by id.
	// If nil - queue not exists
	QueueGet(id uint64) *QueueInfo
	// TaskNew creates new task in queue
	// If nil - queueId invalid
	TaskNew(queueId uint64, parentUUID string, status string, content []byte) *TaskInfo
	// TaskGet gets task by uuid.
	// If nil - task not exists or queueId invalid
	TaskGet(queueId uint64, uuid string) *TaskInfo
	// TasksGet returns all tasks in queue
	TasksGet(queueId uint64) []*TaskInfo
	// TaskStatusSet sets new status with content.
	// If ok==fale - task with uuid not exists or queueId invalid
	TaskStatusSet(queueId uint64, uuid string, status string, content []byte) (ok bool)
	// TaskRemove removes task from queue.
	// If ok==false - task not exists or queueId invalid
	TaskRemove(queueId uint64, uuid string) (ok bool)
}

type StorageWithControl interface {
	Storage
	Start()
	Stop()
}

func New(dbDriver *gorm.DB) StorageWithControl {
	s := &storage{
		db: dbDriver,
	}

	return s
}

type storage struct {
	mu        sync.Mutex
	isStarted bool
	db        *gorm.DB

	cache struct {
		queueByName sync.Map // map[name]*queue
		queueById   sync.Map // map[uint64]*queue
	}
}

func (s *storage) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStarted {
		return
	}

	log.Printf("Storage: starting...\n")

	err := s.db.AutoMigrate(&dbQueue{}, &dbTask{})

	if err != nil {
		panic(fmt.Errorf("auto migrate fail. %s", err))
	}

	s.isStarted = true

	log.Printf("Storage: started\n")
}

func (s *storage) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isStarted {
		return
	}

	log.Printf("SQLiteStorage: stopping...\n")

	s.isStarted = false

	s.cache.queueByName.Range(func(qname, iq interface{}) bool {
		iq.(*queue).stop()
		return true
	})

	s.cache.queueByName = sync.Map{}
	s.cache.queueById = sync.Map{}

	log.Printf("SQLiteStorage: stopped\n")
}

func (s *storage) QueueGetOrCreate(name string) *QueueInfo {
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

		tx := s.db.Begin()

		dbq := &dbQueue{Name: name}
		err := tx.Where(dbq).Take(dbq).Error

		if err != nil {
			if err != gorm.ErrRecordNotFound {
				tx.Rollback()
				panic(fmt.Errorf("get queue `%s` from storage fail. %s", name, err))
			}

			err = tx.Create(dbq).Error

			if err != nil {
				tx.Rollback()
				panic(fmt.Errorf("create queue `%s` fail. %s", name, err))
			}

			tx.Commit()
		}

		q := &queue{dbQueue: dbq, db: s.db}
		q.start()

		s.cache.queueByName.Store(name, q)
		s.cache.queueById.Store(uint64(dbq.ID), q)

		iq = q
		break
	}

	q := iq.(*queue)

	return &QueueInfo{
		Id:   uint64(q.ID),
		Name: q.Name,
	}
}

func (s *storage) queueGet(id uint64) *queue {
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

	return iq.(*queue)
}

func (s *storage) QueueGet(id uint64) (qi *QueueInfo) {
	q := s.queueGet(id)

	if q != nil {
		qi = &QueueInfo{
			Id:   uint64(q.ID),
			Name: q.Name,
		}
	}

	return
}

func (s *storage) TaskNew(queueId uint64, parentUUID string, status string, content []byte) (ti *TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ti = q.taskNew(parentUUID, status, content)
	}

	return
}

func (s *storage) TaskGet(queueId uint64, uuid string) (ti *TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ti = q.taskGetInfo(uuid)
	}

	return
}

func (s *storage) TasksGet(queueId uint64) (ts []*TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ts = q.tasksInfoGet()
	}

	return
}

func (s *storage) TaskStatusSet(queueId uint64, uuid string, status string, content []byte) (ok bool) {
	q := s.queueGet(queueId)

	if q != nil {
		ok = q.taskStatusSet(uuid, status, content)
	}

	return
}

func (s *storage) TaskRemove(queueId uint64, uuid string) (ok bool) {
	q := s.queueGet(queueId)

	if q != nil {
		ok = q.taskRemove(uuid)
	}

	return
}
