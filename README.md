# currency-exchanger

Currency Converter is a Go application that allows users to convert amounts from one currency to another using the latest exchange rates obtained from an external API. It also includes a rate limiter and circuit breaker for enhanced reliability and performance.

## Features

- Convert amounts from one currency to another.
- Automatically updates exchange rates every 3 hours.
- Rate Limiter: Limits the number of API calls to prevent overloading the external service.
- Circuit Breaker: Ensures resilience by temporarily halting requests to the API in case of failure, preventing cascading failures.

## Getting Started

### Prerequisites

- Go 1.13 or higher installed on your machine.
- Internet connection to fetch the latest exchange rates.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/imesh-herath/currency-exchanger.git
   ```
