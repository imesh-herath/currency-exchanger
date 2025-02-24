package server

import (
	"assignment-imesh/configuration"
	"assignment-imesh/http/router"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	httpSrv *http.Server
	wait    time.Duration
}

func NewServer(appConfig configuration.AppConfig) *Server {
	r := router.Init()

	address := "0.0.0.0:" + appConfig.Server.Port
	srv := new(Server)
	srv.httpSrv = &http.Server{
		Addr:         address,
		WriteTimeout: time.Minute * 2,
		ReadTimeout:  time.Minute * 2,
		IdleTimeout:  time.Minute * 2,
		Handler:      r,
	}

	// Expose application metrics
	r.Handle("/metrics", promhttp.Handler())

	return srv
}

func (server *Server) Start() error {
	// Run HTTP server in a goroutine so that it doesn't block
	go func() {
		err := server.httpSrv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("http.server.Init", err)
			panic("HTTP server shutting down unexpectedly...")
		}
	}()

	fmt.Println("http.server.Init", fmt.Sprintf("HTTP server listening on %s", server.httpSrv.Addr))
	return nil
}

func (server *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), server.wait)
	defer cancel()

	err := server.httpSrv.Shutdown(ctx)
	if err != nil {
		fmt.Println("http.server.gracefully.ShutDown", "Unable to stop HTTP server:", err)
	} else {
		fmt.Println("http.server.gracefully.ShutDown", "HTTP server stopped gracefully")
	}
}
