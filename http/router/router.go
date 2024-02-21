package router

import (
	"assignment-imesh/http/controllers"

	"github.com/gorilla/mux"
)

func Init() *mux.Router {
	// cfg := config.App
	router := mux.NewRouter()

	// Set up routes
	router.HandleFunc("/convert", controllers.ConvertCurrencyHandler).Methods("POST")

	// Start the server
	// fmt.Println("Server listening on port 8080...")
	// http.ListenAndServe(":" + cfg.Server.Port, router)

	return router
}
