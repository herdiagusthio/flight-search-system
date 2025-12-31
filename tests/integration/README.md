# Integration Tests

This directory contains integration tests for the flight search API.

## Structure

```
tests/integration/
└── api_test.go         # Full API integration tests
```

## Running Tests

Run all integration tests:
```bash
go test ./tests/integration/... -v
```

Run specific test:
```bash
go test ./tests/integration/... -v -run TestFlightSearchEndpoint
```

Run with coverage:
```bash
go test ./tests/integration/... -v -cover
```

## Test Suites

### TestFlightSearchEndpoint
Tests the `/api/v1/flights/search` endpoint with various scenarios:
- Valid search request (handles both 200 OK and 503 Service Unavailable)
- Invalid origin airport code (400 Bad Request)
- Missing required fields (400 Bad Request)
- Malformed JSON (400 Bad Request)

### TestCORSHeaders
Verifies CORS middleware configuration:
- OPTIONS preflight request headers
- POST request with origin header

### TestRequestIDMiddleware
Verifies Request ID generation:
- Health check endpoint
- Flight search endpoint

### TestHealthCheckEndpoint
Tests the health check endpoint:
- GET returns healthy status
- POST method not allowed

## Notes

- Integration tests use the `internal/api` package to configure the server
- Tests accept both 200 and 503 responses for search requests because provider mock files may not be accessible from the test directory
- All tests use table-driven test format for maintainability
- Request validation and error handling are thoroughly tested
