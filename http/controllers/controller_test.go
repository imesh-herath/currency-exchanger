package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertCurrencyHandler(t *testing.T) {
	// Prepare a sample request
	requestBody := []byte(`{"fromCurrency": "USD", "toCurrency": "EUR", "amount": 100}`)
	req, err := http.NewRequest("POST", "/convert", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ConvertCurrencyHandler)

	// Call the handler function directly and pass in our Request and ResponseRecorder
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the content type header
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned unexpected content type: got %v want %v",
			contentType, expectedContentType)
	}

	// Check the response body
	expectedResponseBody := `{"amount":20,"currency":"EUR"}`
	if rr.Body.String() != expectedResponseBody {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponseBody)
	}
}
