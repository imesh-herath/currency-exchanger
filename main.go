package main

import (
	"assignment-imesh/configuration"
	"assignment-imesh/http/server"
	"assignment-imesh/profiling"
	"assignment-imesh/usecase"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/trace"
	"syscall"
	"time"
)

var (
	httpServer *server.Server
)

func main() {
	// Setup tracing
	tracerCfg := trace.FlightRecorderConfig{
		MinAge:   5 * time.Second,
		MaxBytes: 3 << 20, // 3MB
	}

	// Create and start the FlightRecorder
	fr := trace.NewFlightRecorder(tracerCfg)
	if err := fr.Start(); err != nil {
		panic(err)
	}
	defer fr.Stop()

	// Load configuration from file
	configuration.Init()

	// expose application profiling
	profiling.Profiling(configuration.App)

	httpServer = server.NewServer(configuration.App)

	// Initial update of exchange rates
	err := usecase.UpdateExchangeRates()
	if err != nil {
		fmt.Println("Failed to retrieve initial exchange rates:", err)
		return
	}

	// Start http server
	err = httpServer.Start()
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C), SIGKILL, SIGQUIT or SIGTERM (Ctrl+/)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// Block until we receive our signal
	signal := <-c
	fmt.Println("bootstrap.init.Start", fmt.Sprintf("Received Signal: %s", signal))

	// Dump the FlightRecorder trace to a file on shutdown
	writeTrace(fr)

	// Start destructing the process
	httpServer.Stop()
}

func writeTrace(fr *trace.FlightRecorder) {
	f, _ := os.Create("trace.out")
	defer f.Close()
	fr.WriteTo(f)
}
