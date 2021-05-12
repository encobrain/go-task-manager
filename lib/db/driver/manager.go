package driver

import (
	"fmt"
	"github.com/encobrain/go-task-manager/lib/filepath"
	"github.com/encobrain/go-task-manager/model/config/lib/db/driver"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

func NewManager(config *driver.Manager) Manager {
	return &manager{
		conf:    config,
		drivers: map[string]*gorm.DB{},
	}
}

type manager struct {
	mu      sync.Mutex
	conf    *driver.Manager
	drivers map[string]*gorm.DB
}

func (m *manager) Driver(name string) (driver *gorm.DB, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if drv, ok := m.drivers[name]; ok {
		return drv, nil
	}

	drvConf, ok := m.conf.Drivers[name]

	if !ok {
		err = fmt.Errorf("not found config for driver `%s`", name)
		return
	}

	var dialector gorm.Dialector

	switch drvConf.Db.Type {
	default:
		err = fmt.Errorf("unsupported db.type `%s`", drvConf.Db.Type)
	case "postgres":
		sslMode := "disable"

		if drvConf.Extra["sslEnabled"] == true {
			sslMode = "allow"
		}

		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			drvConf.Connect.Host, drvConf.Connect.Port,
			drvConf.Auth.User, drvConf.Auth.Pass,
			drvConf.Db.Name, sslMode,
		)

		dialector = postgres.Open(dsn)
	case "sqlite":
		dirpath := drvConf.Extra["dirpath"].(string)

		dialector = sqlite.Open(filepath.Resolve(true, dirpath, drvConf.Db.Name))
	}

	dbLog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	driver, err = gorm.Open(dialector, &gorm.Config{
		Logger: dbLog,
	})

	if err != nil {
		return nil, fmt.Errorf("open connection fail. %s", err)
	}

	m.drivers[name] = driver
	return
}
