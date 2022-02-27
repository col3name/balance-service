package server

import (
	"github.com/col3name/balance-transfer/pkg/app/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type HttpServer struct {
	Logger log.Logger
}

func (s *HttpServer) StartServer(port string, handler http.Handler) *http.Server {
	srv := &http.Server{Addr: ":" + port, Handler: handler}
	s.Logger.Error(srv.ListenAndServe())
	return srv
}

func (s *HttpServer) GetKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)

	return osKillSignalChan
}

func (s *HttpServer) WaitForKillSignal(killSignalChan <-chan os.Signal) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		s.Logger.Info("got SIGINT...")
	case syscall.SIGTERM:
		s.Logger.Info("got SIGTERM...")
	}
}
