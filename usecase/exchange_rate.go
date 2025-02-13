package usecase

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

var (
	lastUpdated   time.Time
	exchangeRates map[string]float64
	lock          sync.Mutex
	cb            *gobreaker.CircuitBreaker
)

func init() {
	log.Println("Initializing Circuit Breaker...")
	cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ExchangeRate",
		MaxRequests: 5,
		Interval:    0,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the circuit breaker if failures exceed 5
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit Breaker %s changed from %s to %s\n", name, from, to)
		},
	})
	log.Println("Circuit Breaker initialized successfully!")
}

func GetExchangeRateWithCircuitBreaker(fromCurrency, toCurrency string) (float64, error) {
	log.Printf("Attempting to get exchange rate for %s to %s using Circuit Breaker...\n", fromCurrency, toCurrency)
	result, err := cb.Execute(func() (interface{}, error) {
		return GetExchangeRate(fromCurrency, toCurrency)
	})
	if err != nil {
		log.Printf("Circuit Breaker rejected the request or an error occurred: %s\n", err)
		return 0, err
	}
	log.Printf("Exchange rate retrieved successfully for %s to %s: %f\n", fromCurrency, toCurrency, result.(float64))
	return result.(float64), nil
}

func GetExchangeRate(fromCurrency string, toCurrency string) (float64, error) {
	lock.Lock()
	defer lock.Unlock()

	// Check if the last updated currency rate map is greater than or equal to 3 hours
	if time.Since(lastUpdated).Hours() >= 3 || exchangeRates == nil {
		err := UpdateExchangeRates()
		if err != nil {
			return 0, err
		}
	}

	// Set the above retrieved rate to the from currency
	rate, ok := exchangeRates[fromCurrency]
	if !ok {
		return 0, fmt.Errorf("exchange rate not available for %s", fromCurrency)
	}

	return rate, nil
}
