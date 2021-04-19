package driver

import "github.com/encobrain/go-task-manager/model/config/lib/db"

type Manager struct {
	Drivers map[string]db.Driver
}

func (Manager) Pathfile() string {
	return "db.driver.manager.yaml"
}
