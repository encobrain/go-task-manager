package config

import (
	"github.com/encobrain/go-task-manager/lib/filepath"
	"runtime"
)

type Process struct {
	Name string `yaml:"name"   env:"NAME"   long:"name"   description:"Process name"`
	Run  struct {
		PidPathfile string `yaml:"pid_pathfile"   env:"RUN_PID_PATHFILE"   long:"run.pidPathfile"   description:"Pathfile for store process id" default:"./run.pid"`
		Debug       struct {
			Pprof       bool `yaml:"pprof" env:"RUN_DEBUG_PPROF" long:"run.debug.pprof" description:"Enable debug pprof"`
			CtxInterval int  `yaml:"ctx_interval" env:"RUN_DEBUG_CTX_INTERVAL" long:"run.debug.ctxInterval" description:"Show contexts every X sec. 0 - disable" default:"0"`
		}
	} `yaml:"run" namespace:"run"`
	Threads struct {
		Max int `yaml:"max"   env:"THREADS_MAX"   long:"threads.max"   description:"Use max threads. 0=cpu count"   default:"0"`
	} `yaml:"threads"`
}

func (p Process) Pathfile() string {
	return "process.yaml"
}

func (p *Process) Init() {
	p.Run.PidPathfile = filepath.Resolve(true, p.Run.PidPathfile)
}

func (p *Process) Apply() {
	c := p.Threads.Max
	if c == 0 {
		c = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(c)
}
