package storage

import "github.com/encobrain/go-task-manager/lib/filepath"

type SQLite struct {
	Db struct {
		Path string `yaml:"path" long:"db.path" description:"Path to storage db" default:"./storage.db.sqlite3"`
	} `yaml:"db"`
}

func (s *SQLite) Init() {
	s.Db.Path = filepath.Resolve(true, s.Db.Path)
}
