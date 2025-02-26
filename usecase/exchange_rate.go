package usecase

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

var (
	lastUpdated         time.Time
	exchangeRates       map[string]float64
	lock                sync.Mutex
	latencyWindow       []time.Duration
	latencyLock         sync.Mutex
	percentileThreshold = 800 * time.Millisecond // 300ms threshold
	maxWindowSize       = 100                    // Sliding window size

	// Circuit breaker states
	circuitOpen      = false
	lastOpenedTime   time.Time
	circuitDuration  = 5 * time.Second // Open state duration
	halfOpen         = false
	testRequestCount = 0
	maxTestRequests  = 3 // Number of requests in Half-Open state before deciding
)

func init() {
	log.Println("Initializing Hystrix Circuit Breaker...")

	hystrix.ConfigureCommand("exchangeRate", hystrix.CommandConfig{
		Timeout:               10000, // Timeout in milliseconds
		MaxConcurrentRequests: 10,    // Limit concurrent requests
		ErrorPercentThreshold: 100,   // Disable error-based trips
		SleepWindow:           5000,  // Circuit breaker resets after 5s
	})

	log.Println("Hystrix Circuit Breaker initialized successfully!")
}

func GetExchangeRateWithCircuitBreaker(fromCurrency, toCurrency string) (float64, error) {
	log.Printf("Attempting to get exchange rate for %s to %s using Circuit Breaker...\n", fromCurrency, toCurrency)

	// **OPEN STATE**: Reject requests if circuit is open
	if circuitOpen {
		if time.Since(lastOpenedTime) > circuitDuration {
			log.Println("Transitioning to HALF-OPEN state: Allowing test requests.")
			circuitOpen = false
			halfOpen = true
			testRequestCount = 0
		} else {
			log.Println("Circuit is open. Rejecting request.")
			resetLatencyTracking()
			return 0, fmt.Errorf("circuit breaker open, reset latency tracking")
		}
	}

	// **HALF-OPEN STATE**: Allow limited test requests
	if halfOpen {
		if testRequestCount >= maxTestRequests {
			log.Println("Half-Open test failed. Returning to Open state.")
			circuitOpen = true
			lastOpenedTime = time.Now()
			halfOpen = false
			return 0, fmt.Errorf("circuit breaker open (after half-open test)")
		}
		testRequestCount++
		log.Printf("Half-Open Test: Attempt #%d\n", testRequestCount)
	}

	// **Check Latency Before Executing Request**
	percentileLatency := getLatencyPercentile(90)
	log.Printf("Current 90th percentile latency: %.2f ms", percentileLatency.Seconds()*1000)

	if percentileLatency > percentileThreshold {
		log.Printf("Tripping circuit due to high latency: %.2f ms > %.2f ms", percentileLatency.Seconds()*1000, percentileThreshold.Seconds()*1000)
		circuitOpen = true
		lastOpenedTime = time.Now()
		halfOpen = false
		return 0, fmt.Errorf("circuit breaker manually opened due to high latency")
	}

	var exchangeRate float64
	startTime := time.Now()

	if rand.Intn(10) == 0 { // 33% chance of failure
		// Simulate long delay in case of failure
		time.Sleep(time.Duration(500+rand.Intn(600)) * time.Millisecond) // Simulated failure delay
	}

	time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

	// **Execute GetExchangeRate Inside Circuit Breaker**
	err := hystrix.Do("exchangeRate", func() error {

		// // Simulate different request latencies
		// randomDelay := time.Duration(2000+rand.Intn(1000)) * time.Millisecond // Random 2000-3000ms delay
		// log.Printf("Simulating random delay: %.2f ms", randomDelay.Seconds()*1000)
		// time.Sleep(randomDelay)

		// Measure actual API latency
		rate, err := GetExchangeRate(fromCurrency, toCurrency)
		if err != nil {
			return err
		}

		latency := time.Since(startTime)
		log.Printf("Request latency: %.2f ms, Latency Window: %d\n", latency.Seconds()*1000, latencyWindow)

		// Store the latest latency in the sliding window
		recordLatency(latency)

		exchangeRate = rate
		return nil
	}, nil)

	if err != nil {
		log.Printf("Circuit Breaker rejected the request or an error occurred: %s\n", err)
		return 0, err
	}

	// **HALF-OPEN STATE SUCCESS: Move to CLOSED**
	if halfOpen {
		log.Println("Half-Open Test Succeeded. Closing the circuit.")
		halfOpen = false
		circuitOpen = false
		resetLatencyTracking()
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
	sortedLatencies := append([]time.Duration(nil), latencyWindow...)
	sort.Slice(sortedLatencies, func(i, j int) bool { return sortedLatencies[i] < sortedLatencies[j] })

	// Handle edge cases where window size is small
	if len(sortedLatencies) == 1 {
		log.Printf("Single latency recorded, using it for percentile.")
		return sortedLatencies[0]
	}

	// Calculate index for percentile
	index := (percentile * len(sortedLatencies)) / 100
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}

	// Log percentile latency
	log.Printf("90th percentile latency: %.2f ms", sortedLatencies[index].Seconds()*1000)
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
