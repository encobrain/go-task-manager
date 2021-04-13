package main

import (
	"fmt"
	"github.com/encobrain/go-context.v2"
	"github.com/encobrain/go-task-manager/cli"
	contextSupp "github.com/encobrain/go-task-manager/lib/context"
	"github.com/encobrain/go-task-manager/lib/storage/queue"
	"github.com/encobrain/go-task-manager/lib/storage/sqlite"
	"github.com/encobrain/go-task-manager/model/config/service"
	task_manager "github.com/encobrain/go-task-manager/service/task-manager"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type CmdStart struct {
	cli.CmdStart
}

func (c CmdStart) Execute(args []string) (err error) {
	conf.Init()
	success, err := c.CmdStart.Execute(conf.Process.Run.PidPathfile)

	if err != nil || !success {
		return fmt.Errorf("execute start fail. %s", err)
	}

	startCtx := context.Main.Child("start", c.start).Go()

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
		case <-startCtx.Done():
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

func (c CmdStart) panicHandler(ctx context.Context, panicErr interface{}) {
	log.Printf("Service panic. %s\n", panicErr)

	ctx.Cancel(fmt.Errorf("service panic"))
}

func (c CmdStart) start(ctx context.Context) {
	ctx.PanicHandlerSet(c.panicHandler)

	serverCtx := ctx.Child("server", c.startServer).
		ValueSet("config", &service.TaskManager{})

	stor := sqlite.New(&conf.Storage)
	sqm := queue.NewManager(stor)

	ctx.ValueSet("storage.queue.manager", sqm)

	stor.Start()
	serverCtx.Go()

	<-ctx.Done()

	log.Printf("Stopping service...\n")

	stor.Stop()
}

func (c CmdStart) startServer(ctx context.Context) {
	tmService := task_manager.New(ctx)
	tmService.Start()

	serveMux := http.NewServeMux()
	upgrader := websocket.Upgrader{}

	serveMux.HandleFunc(conf.Server.Listen.Path, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New connection\n")

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

	addr := fmt.Sprintf("%s:%d", conf.Server.Listen.Ip, conf.Server.Listen.Port)

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

	<-ctx.Done()

	log.Printf("Closing HTTP server...\n")

	err := server.Close()

	if err != nil {
		log.Printf("Close HTTP server fail. %s\n", err)
	}

}
