package entities


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