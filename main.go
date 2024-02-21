package main

import (
	"assignment-imesh/configuration"
	"assignment-imesh/http/controllers"
	"assignment-imesh/http/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	httpServer *server.Server
)

func main() {
	// Load configuration from file
	configuration.Init()

	httpServer = server.NewServer(configuration.App)

	// Initial update of exchange rates
	err := controllers.UpdateExchangeRates()
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

	// Start destructing the process
	httpServer.Stop()
}
