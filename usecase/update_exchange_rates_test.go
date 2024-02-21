package usecase

import (
    "assignment-imesh/configuration"
    "assignment-imesh/entities"
    // "bytes"
    "encoding/json"
    // "io/ioutil"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestUpdateExchangeRates(t *testing.T) {
    // Mock the configuration URL
    configuration.App.ExchangeRateConfig.URL = "http://example.com/exchange-rates"

    // Mock the exchange rates response
    exchangeRatesResponse := entities.ExchangeRatesResponse{
        Rates: map[string]float64{
            "USD": 1.0,
            "LKR": 311.758,
            "GBP": 0.73,
        },
    }

    // Convert exchange rates response to JSON
    jsonExchangeRates, err := json.Marshal(exchangeRatesResponse)
    if err != nil {
        t.Errorf("failed to marshal exchange rates response: %v", err)
    }

    // Create a mock HTTP server
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write(jsonExchangeRates)
    }))
    defer mockServer.Close()

    // Replace the configuration URL with the mock server's URL
    configuration.App.ExchangeRateConfig.URL = mockServer.URL

    // Invoke the function to be tested
    err = UpdateExchangeRates()
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Validate if exchange rates were updated
    if len(exchangeRates) != len(exchangeRatesResponse.Rates) {
        t.Errorf("exchange rates not updated correctly")
    }

    // Validate if lastUpdated time was updated
    if time.Since(lastUpdated) > time.Second {
        t.Errorf("lastUpdated time not updated")
    }
}
