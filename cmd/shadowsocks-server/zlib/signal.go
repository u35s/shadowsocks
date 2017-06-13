package zlib

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/u35s/glog"
)

func WaitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	for sig := range sigChan {
		if sig == syscall.SIGHUP {
		} else {
			// is this going to happen?
			glog.Inf("[exit],%v,caught signal", sig)
			os.Exit(0)
		}
	}
}
