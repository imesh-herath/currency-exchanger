package usecase

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

var (
	lastUpdated   time.Time
	exchangeRates map[string]float64
	lock          sync.Mutex
	cb            *gobreaker.CircuitBreaker
	latencyWindow []time.Duration
	latencyLock   sync.Mutex
)

const (
	percentileThreshold = 600 * time.Millisecond // 600ms threshold
	maxWindowSize       = 100                    // Track the last 5 latencies
)

func init() {
	log.Println("Initializing Circuit Breaker...")
	cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ExchangeRate",
		MaxRequests: 5,
		Interval:    0,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			log.Printf("ReadyToTrip invoked. ConsecutiveFailures: %d", counts.ConsecutiveFailures)

			// Check if the 90th percentile latency is above the threshold
			percentileLatency := getLatencyPercentile(90)
			log.Printf("Current 90th percentile latency: %.2f ms", percentileLatency.Seconds()*1000)

			// Trip if failures exceed the threshold or latency exceeds the percentile threshold
			return percentileLatency > percentileThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("State Change Detected: Circuit Breaker %s changed from %s to %s\n", name, from, to)

			// Log the state explicitly
			log.Printf("Current State: %s", to)

			// Checking if state has changed to CLOSED, OPEN, or HALF-OPEN
			if to == gobreaker.StateClosed {
				log.Println("Circuit Breaker state is now CLOSED")
				resetLatencyTracking() // Reset latency window on circuit resetS
			} else if to == gobreaker.StateOpen {
				log.Println("Circuit Breaker state is now OPEN")
			} else if to == gobreaker.StateHalfOpen {
				log.Println("Circuit Breaker state is now HALF-OPEN")
			} else {
				log.Println("Unexpected Circuit Breaker state change detected.")
			}
		},
	})
	log.Println("Circuit Breaker initialized successfully!")
}

func GetExchangeRateWithCircuitBreaker(fromCurrency, toCurrency string) (float64, error) {
	log.Printf("Attempting to get exchange rate for %s to %s using Circuit Breaker...\n", fromCurrency, toCurrency)

	startTime := time.Now()

	result, err := cb.Execute(func() (interface{}, error) {
		if rand.Intn(10) == 0 { // 33% chance of failure
			// Simulate long delay in case of failure
			time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond) // Simulated failure delay
			return nil, errors.New("simulated failure")
		}

		time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

		return GetExchangeRate(fromCurrency, toCurrency)
	})

	latency := time.Since(startTime)

	// Store the latest latency in the sliding window
	if cb.State() != gobreaker.StateOpen {
		recordLatency(latency)
	}

	if err != nil {
		log.Printf("Circuit Breaker rejected the request or an error occurred: %s\n", err)
		return 0, err
	}

	// **Ensure the result is not nil before asserting**
	if result == nil {
		log.Println("Received nil exchange rate result, returning error")
		return 0, errors.New("unexpected nil result from exchange rate fetch")
	}

	exchangeRate, ok := result.(float64)
	if !ok {
		log.Println("Unexpected result type from exchange rate fetch")
		return 0, errors.New("unexpected result type")
	}

	log.Printf("Exchange rate retrieved successfully for %s to %s: %f\n", fromCurrency, toCurrency, exchangeRate)
	return exchangeRate, nil
}

func recordLatency(latency time.Duration) {
	latencyLock.Lock()
	defer latencyLock.Unlock()

	latencyWindow = append(latencyWindow, latency)
	if len(latencyWindow) > maxWindowSize {
		latencyWindow = latencyWindow[1:] // If the number of stored latencies exceeds maxWindowSize, it removes the oldest entry (first element).
	}

	log.Printf("Recorded latency: %.2f ms (window size: %d)", latency.Seconds()*1000, len(latencyWindow))
}

func getLatencyPercentile(percentile int) time.Duration {
	latencyLock.Lock()
	defer latencyLock.Unlock()

	if len(latencyWindow) == 0 {
		log.Println("No latencies recorded yet. Returning default latency.")
		return 200 * time.Millisecond
	}

	// Copy and sort latencies
	/*
		Percentile calculations requires data to be sorted in ascending order to identify the
		correct value at the given percentile.
	*/
	sortedLatencies := append([]time.Duration(nil), latencyWindow...)
	sort.Slice(sortedLatencies, func(i, j int) bool { return sortedLatencies[i] < sortedLatencies[j] })

	// Get percentile index
	/*
		Used this formula because it scales the percentile (0 to 100) to the length of the sorted latencies slice.
		This way, ensure that the index is within the bounds of the slice.
		If the index is greater than the length of the slice, then return the last element.
	*/
	index := (percentile * len(sortedLatencies)) / 100
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}

	log.Printf("Latency window: %v", sortedLatencies)
	log.Printf("%dth percentile latency: %.2f ms", percentile, sortedLatencies[index].Seconds()*1000)

	return sortedLatencies[index]
}

func resetLatencyTracking() {
	latencyLock.Lock()
	defer latencyLock.Unlock()
	latencyWindow = nil
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
