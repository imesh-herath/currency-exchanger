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
	//Set a mutex lock for syncrization for prevent race condition when access to the function
	lock.Lock()
	defer lock.Unlock() //unlock the end of the process

	//Checking wheather last updated currency rate map is greater or equal to 3 hours
	if time.Since(lastUpdated).Hours() >= 3 || exchangeRates == nil {
		err := UpdateExchangeRates()
		if err != nil {
			return 0, err
		}
	}

	//Set the above retrived rate to the from currency
	rate, ok := exchangeRates[fromCurrency]
	if !ok {
		return 0, fmt.Errorf("exchange rate not available for %s", fromCurrency)
	}

	return rate, nil
}
