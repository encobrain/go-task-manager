package sqlite

import (
	"fmt"
	"github.com/encobrain/go-task-manager/lib/storage"
	confStorage "github.com/encobrain/go-task-manager/model/config/lib/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
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

	dbLog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(s.conf.Db.Path), &gorm.Config{
		Logger: dbLog,
	})

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

	s.cache.queueByName.Range(func(qname, iq interface{}) bool {
		iq.(*queue).stop()
		return true
	})

	s.cache.queueByName = sync.Map{}
	s.cache.queueById = sync.Map{}

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

func (s *Storage) queueGet(id uint64) *queue {
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

func (s *Storage) QueueGet(id uint64) (qi *storage.QueueInfo) {
	q := s.queueGet(id)

	if q != nil {
		qi = &storage.QueueInfo{
			Id:   uint64(q.ID),
			Name: q.Name,
		}
	}

	return
}

func (s *Storage) TaskNew(queueId uint64, parentUUID string, status string, content []byte) (ti *storage.TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ti = q.taskNew(parentUUID, status, content)
	}

	return
}

func (s *Storage) TaskGet(queueId uint64, uuid string) (ti *storage.TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ti = q.taskGetInfo(uuid)
	}

	return
}

func (s *Storage) TasksGet(queueId uint64) (ts []*storage.TaskInfo) {
	q := s.queueGet(queueId)

	if q != nil {
		ts = q.tasksInfoGet()
	}

	return
}

func (s *Storage) TaskStatusSet(queueId uint64, uuid string, status string, content []byte) (ok bool) {
	q := s.queueGet(queueId)

	if q != nil {
		ok = q.taskStatusSet(uuid, status, content)
	}

	return
}

func (s *Storage) TaskRemove(queueId uint64, uuid string) (ok bool) {
	q := s.queueGet(queueId)

	if q != nil {
		ok = q.taskRemove(uuid)
	}

	return
}
