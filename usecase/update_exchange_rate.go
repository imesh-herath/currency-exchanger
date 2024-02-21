package usecase

import (
	"assignment-imesh/configuration"
	"assignment-imesh/entities"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

	var exchangeRatesResponse entities.ExchangeRatesResponse
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