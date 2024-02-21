package usecase

import (
	"testing"
	"time"
)

func TestGetExchangeRate(t *testing.T) {
	// Mock the lastUpdated time to be three hours ago
	lastUpdated = time.Now().Add(-3 * time.Hour)

	// Mock exchange rates
	exchangeRates = map[string]float64{
		"USD": 1.0,
		"LKR": 314.0,
		"GBP": 0.73,
	}

	// Test retrieving exchange rate for existing currency
	expectedRate := 1.0
	rate, err := GetExchangeRate("LKR", "USD")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rate != expectedRate {
		t.Errorf("unexpected rate: got %f, want %f", rate, expectedRate)
	}

	// Test retrieving exchange rate for non-existing currency
	_, err = GetExchangeRate("XYZ", "ABC")
	if err == nil {
		t.Error("expected error for non-existing currency, but got nil")
	}
}
