package driver

import "gorm.io/gorm"

type Manager interface {
	Driver(name string) (driver *gorm.DB, err error)
}
