package cmd

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/cli"
	contextSupp "github.com/encobrain/go-task-manager/lib/context"
	"github.com/encobrain/go-task-manager/lib/db/driver"
	"github.com/encobrain/go-task-manager/lib/db/storage"
	"github.com/encobrain/go-task-manager/lib/storage/queue"
	confService "github.com/encobrain/go-task-manager/model/config/service"
	"github.com/encobrain/go-task-manager/service"
	"github.com/encobrain/go-task-manager/service/task-manager"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"time"
)

type Start struct {
	*Config      `no-flag:"true"`
	cli.CmdStart `no-flag:"true"`
}

func (c Start) Execute(args []string) (err error) {
	c.Config.Init()
	success, err := c.CmdStart.Execute(c.Config.Process.Run.PidPathfile)

	if err != nil || !success {
		return fmt.Errorf("execute start fail. %s", err)
	}

	var serveMux = http.NewServeMux()

	if c.Config.Process.Run.Debug.Pprof {
		serveMux.HandleFunc("/debug/pprof/", pprof.Index)
		serveMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		serveMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		serveMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		serveMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		log.Printf("DEBUG: added pprof")
	}

	var debugShowContexts <-chan time.Time

	if c.Config.Process.Run.Debug.CtxInterval > 0 {
		debugShowContexts = time.Tick(time.Second * time.Duration(c.Config.Process.Run.Debug.CtxInterval))

		log.Printf("DEBUG: added show contexts every %d sec", c.Config.Process.Run.Debug.CtxInterval)
	}

	startCtx := context.Main.Child("start", c.start).
		ValueSet("serveMux", serveMux).Go()

	go func() {
		sig := <-c.CmdStart.ShutdownSignals()

		log.Printf("Received signal: %s\n", sig.String())

		startCtx.Cancel(fmt.Errorf("service shutdown"))
	}()

	var prevRuns string

	for {
		select {
		case <-context.Main.ChildsFinished(true):
			log.Printf("Service stopped\n")
			return
		case <-debugShowContexts:
			currRuns := contextSupp.GetDeadlocksInfo(context.Main)

			changed := ""

			if currRuns != prevRuns {
				changed = " (changed)"
			}

			prevRuns = currRuns

			log.Printf("Working contexts%s:\n%s\n", changed, currRuns)
		case <-startCtx.Done():
			debugShowContexts = nil

			select {
			case <-context.Main.ChildsFinished(true):
				continue
			case <-time.After(time.Second):
				currRuns := contextSupp.GetDeadlocksInfo(context.Main)

				if currRuns != prevRuns {
					log.Printf("Waiting for close contexts:\n%s\n", currRuns)
					prevRuns = currRuns
				}
			}
		}
	}
}

func (c Start) panicHandler(ctx context.Context, panicErr interface{}) {
	log.Printf("Service panic. %s\n", panicErr)
	debug.PrintStack()

	ctx.Cancel(fmt.Errorf("service panic"))
}

// ctx should contain vars:
//   serveMux *http.ServeMux
func (c Start) start(ctx context.Context) {
	ctx.PanicHandlerSet(c.panicHandler)

	dbDrvMng := driver.NewManager(&c.Config.DbDriverManager)
	dbDrv, err := dbDrvMng.Driver(c.Config.Storage.Db.DriverId)

	if err != nil {
		panic(fmt.Errorf("get db driver `%s` fail. %s", c.Config.Storage.Db.DriverId, err))
	}

	serverCtx := ctx.Child("server", c.startServer).
		ValueSet("config", &confService.TaskManager{})

	stor := storage.New(dbDrv)
	sqm := queue.NewManager(stor)

	ctx.ValueSet("storage.queue.manager", sqm)

	stor.Start()
	serverCtx.Go()

	select {
	case <-ctx.Done():
	case <-serverCtx.Finished(false):
	}

	log.Printf("Stopping storage...")

	stor.Stop()
}

// ctx should contain vars:
//   serveMux *http.ServeMux
func (c Start) startServer(ctx context.Context) {
	serveMux := ctx.Value("serveMux").(*http.ServeMux)

	tmService := task_manager.New(ctx)
	tmService.Start()

	upgrader := websocket.Upgrader{}

	serveMux.HandleFunc(c.Config.Server.Listen.Path, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New connection from %s\n", r.RemoteAddr)

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Printf("Upgrade connection fail. %s\n", err)
			w.WriteHeader(http.StatusUpgradeRequired)
			_, err = w.Write([]byte("Upgrade connection fail"))
		} else {
			err = tmService.ConnServe(conn)

			if err != nil {
				log.Printf("Serve connection fail. %s\n", err)

				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte("Serve connection fail"))
			}
		}

		if err != nil {
			log.Printf("Write response fail. %s\n", err)
		}
	})

	addr := fmt.Sprintf("%s:%d", c.Config.Server.Listen.Ip, c.Config.Server.Listen.Port)

	server := http.Server{
		Addr:    addr,
		Handler: serveMux,
	}

	ctx.Child("listen", func(_ context.Context) {
		log.Printf("HTTP Server started listening on %s\n", addr)

		err := server.ListenAndServe()

		if err == http.ErrServerClosed {
			log.Printf("HTTP server closed\n")

			return
		}

		panic(fmt.Errorf("server listen fail. %s", err))
	}).Go()

	select {
	case <-ctx.Done():
	case <-tmService.StatusWait(service.StatusStopped):
		ctx.Cancel(fmt.Errorf("task manager stopped"))
	}

	log.Printf("Closing HTTP server because: %s ...\n", ctx.Err())

	err := server.Close()

	if err != nil {
		log.Printf("Close HTTP server fail. %s\n", err)
	}

}
