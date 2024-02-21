package usecase

import (
	"fmt"
	"sync"
	"time"
)

var (
	lastUpdated   time.Time
	exchangeRates map[string]float64
	lock          sync.Mutex
)

func GetExchangeRate(fromCurrency string, toCurrency string) (float64, error) {
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
