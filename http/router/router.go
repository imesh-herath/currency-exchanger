package router

import (
	"assignment-imesh/http/controllers"

	"github.com/gorilla/mux"
)

func Init() *mux.Router {
	router := mux.NewRouter()

	// Set up routes
	router.HandleFunc("/convert", controllers.ConvertCurrencyHandler).Methods("GET")

	return router
}
