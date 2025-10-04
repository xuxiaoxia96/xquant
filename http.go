package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"

	"xquant/pkg/log"
	"xquant/pkg/utils"
)

var finalizeOnce sync.Once

// finalize execute the custom finalize logic once before exit.
// your program global clean logic should be placed within finalizeOnce.
func finalize() {
	finalizeOnce.Do(func() {
		log.Infof("do finalize")
	})
}

func customWaitSignal(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)

	select {
	case sig := <-signals:
		log.Infof("Received signal: %s\n", sig)
		finalize()
		// graceful shutdown
	case err := <-errCh:
		// error occurs, exit immediately
		return err
	}
	return nil
}

func runHTTP() {
	// normal http server
	port := utils.GetEnvWithDefault("XQUANT_PORT_OF_HTTP", "8889")
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%s", port)),
		server.WithSenseClientDisconnection(true),
		server.WithMaxRequestBodySize(64*1024*1024), // set max request body size to 64MB for 2M token length
		server.WithExitWaitTime(getExitWaitTime()),
	)
	h.NoHijackConnPool = true
	h.SetCustomSignalWaiter(customWaitSignal)
	register(h)
	registerMw(h)
	h.Spin()
}

func registerMw(h *server.Hertz) {
	// recovery
	h.Use(recovery.Recovery())

	// gzip
	h.Use(gzip.Gzip(gzip.DefaultCompression))

	// access log
	//h.Use(middlewares.AccessLog())

	// cors
	//h.Use(cors.Default())
}

func getExitWaitTime() time.Duration {
	exitWaitTime := utils.GetEnvWithDefault("XQUANT_EXIT_WAIT_TIME_SECONDS", "110")
	exitWaitTimeInt, err := strconv.Atoi(exitWaitTime)
	if err != nil {
		return 110 * time.Second
	}
	return time.Duration(exitWaitTimeInt) * time.Second
}
