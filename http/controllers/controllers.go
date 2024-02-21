package controllers

import (
	"assignment-imesh/entities"
	"assignment-imesh/usecase"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)


func ConvertCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	var req entities.ConvertRequest

	//Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %s\n", err)
		return
	}

	//Unmarshal the request body
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Printf("Failed to unmarshal request: %s\n", err)
		return
	}

	//Get the Exchange rates for given currencies
	exchangeRate, err := usecase.GetExchangeRate(req.FromCurrency, req.ToCurrency)
	if err != nil {
		log.Printf("Failed to retrieve exchange rate: %s\n", err)
		return
	}

	convertedAmount := req.Amount / exchangeRate

	res := entities.ConvertResponse{
		Amount:   convertedAmount,
		Currency: req.ToCurrency,
	}

	// Marshal the response struct into JSON
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		log.Printf("Failed to marshal response, %s\n", err)
		return
	}

	// Write the JSON response to the response writer
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Printf("Failed to write response, %s\n", err)
		return
	}

	fmt.Println("=============================================================================================")
	log.Printf("Conversion successful: %f %s converted to %f %s\n", req.Amount, req.FromCurrency, convertedAmount, req.ToCurrency)
	fmt.Println("=============================================================================================")
}
