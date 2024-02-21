package controllers

import (
	"assignment-imesh/configuration"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type ExchangeRatesResponse struct {
	Rates map[string]float64 `json:"conversion_rates"`
}

type ConvertRequest struct {
	FromCurrency string  `json:"fromCurrency"`
	Amount       float64 `json:"amount"`
	ToCurrency   string  `json:"toCurrency"`
}

type ConvertResponse struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

var (
	lastUpdated   time.Time
	exchangeRates map[string]float64
	lock          sync.Mutex
)

// UpdateExchangeRates fetches the latest exchange rates from the API
func UpdateExchangeRates() error {
	url := configuration.App.ExchangeRateConfig.URL
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch exchange rates: %s\n", err)
		return err
	}
	defer resp.Body.Close()

	var exchangeRatesResponse ExchangeRatesResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %s\n", err)
		return err
	}

	err = json.Unmarshal(body, &exchangeRatesResponse)
	if err != nil {
		log.Printf("Failed to unmarshal exchange rates: %s\n", err)
		return err
	}

	lock.Lock()
	exchangeRates = exchangeRatesResponse.Rates
	lastUpdated = time.Now()
	lock.Unlock()

	log.Println("Exchange rates sucessfully updated")

	return nil
}

func getExchangeRate(fromCurrency string, toCurrency string) (float64, error) {
	lock.Lock()
	defer lock.Unlock()

	if time.Since(lastUpdated).Hours() >= 3 || exchangeRates == nil {
		err := UpdateExchangeRates()
		if err != nil {
			return 0, err
		}
	}

	rate, ok := exchangeRates[fromCurrency]
	if !ok {
		return 0, fmt.Errorf("exchange rate not available for %s", fromCurrency)
	}

	return rate, nil
}

func ConvertCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	var req ConvertRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %s\n", err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Printf("Failed to unmarshal request: %s\n", err)
		return
	}

	exchangeRate, err := getExchangeRate(req.FromCurrency, req.ToCurrency)
	if err != nil {
		log.Printf("Failed to retrieve exchange rate: %s\n", err)
		return
	}

	convertedAmount := req.Amount / exchangeRate

	res := ConvertResponse{
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
