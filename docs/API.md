# API Documentation

This document provides comprehensive information about the Flight Search API endpoints, request/response schemas, and usage examples.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
- [Request/Response Examples](#requestresponse-examples)
- [Error Handling](#error-handling)
- [Swagger UI](#swagger-ui)

## Overview

The Flight Search API provides a unified interface to search and aggregate flight information from multiple Indonesian airline providers including:
- Garuda Indonesia
- Lion Air
- Batik Air
- AirAsia

The API follows RESTful principles and returns JSON responses.

## Base URL

```
Development: http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. Future versions may implement API key authentication.

## Endpoints

### Health Check

Check if the API service is running and healthy.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy"
}
```

### Search Flights

Search for available flights based on criteria.

**Endpoint:** `POST /api/v1/flights/search`

**Request Body:**

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `origin` | string | Yes | Origin airport IATA code (3 letters) | `"CGK"` |
| `destination` | string | Yes | Destination airport IATA code (3 letters) | `"DPS"` |
| `departureDate` | string | Yes | Departure date in YYYY-MM-DD format | `"2025-12-15"` |
| `passengers` | integer | Yes | Number of passengers (1-9) | `1` |
| `class` | string | No | Cabin class: economy, business, first | `"economy"` |
| `sortBy` | string | No | Sort order: best, price, duration, departure | `"price"` |
| `filters` | object | No | Optional filters | See below |

**Filters Object:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `maxPrice` | number | Maximum price in IDR | `5000000` |
| `maxStops` | integer | Maximum number of stops | `1` |
| `airlines` | array | Filter by airline codes | `["GA", "JT"]` |
| `departureTimeRange` | object | Departure time range (HH:MM) | `{"start": "06:00", "end": "22:00"}` |
| `arrivalTimeRange` | object | Arrival time range (HH:MM) | `{"start": "06:00", "end": "22:00"}` |
| `durationRange` | object | Flight duration range in minutes | `{"minMinutes": 60, "maxMinutes": 300}` |

**Response:**

```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 15,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 1234,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "QZ520_AirAsia",
      "provider": "AirAsia",
      "airline": {
        "name": "AirAsia",
        "code": "QZ"
      },
      "flight_number": "QZ520",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T04:45:00+07:00",
        "timestamp": 1734213900
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T07:25:00+08:00",
        "timestamp": 1734223500
      },
      "duration": {
        "total_minutes": 100,
        "formatted": "1h 40m"
      },
      "stops": 0,
      "price": {
        "amount": 650000,
        "currency": "IDR",
        "formatted": "Rp 650.000"
      },
      "available_seats": 67,
      "cabin_class": "economy",
      "aircraft": "Airbus A320",
      "amenities": [],
      "baggage": {
        "carry_on": "7 kg",
        "checked": "Checked bags additional fee"
      }
    },
    {
      "id": "GA400_GarudaIndonesia",
      "provider": "Garuda Indonesia",
      "airline": {
        "name": "Garuda Indonesia",
        "code": "GA"
      },
      "flight_number": "GA400",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T06:00:00+07:00",
        "timestamp": 1734218400
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T08:50:00+08:00",
        "timestamp": 1734228600
      },
      "duration": {
        "total_minutes": 110,
        "formatted": "1h 50m"
      },
      "stops": 0,
      "price": {
        "amount": 1250000,
        "currency": "IDR",
        "formatted": "Rp 1.250.000"
      },
      "available_seats": 28,
      "cabin_class": "economy",
      "aircraft": "Boeing 737-800",
      "amenities": ["wifi", "meal", "entertainment"],
      "baggage": {
        "carry_on": "7 kg",
        "checked": "20 kg"
      }
    }
  ]
}
```

## Request/Response Examples

### Example 1: Basic Search

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response:** 200 OK with flight results (approximately 12 flights from all providers)

### Example 2: Search with Price and Time Filters

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy",
    "sortBy": "price",
    "filters": {
      "maxPrice": 1000000,
      "maxStops": 0,
      "departureTimeRange": {
        "start": "06:00",
        "end": "12:00"
      }
    }
  }'
```

**Response:** 200 OK with filtered flights (budget flights in morning time slot)

### Example 3: Search with Duration and Airline Filters

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 2,
    "class": "economy",
    "sortBy": "duration",
    "filters": {
      "durationRange": {
        "minMinutes": 90,
        "maxMinutes": 120
      },
      "airlines": ["GA", "QZ"]
    }
  }'
```

**Response:** 200 OK with flights from Garuda Indonesia (GA) and AirAsia (QZ) with 90-120 minute duration

## Error Handling

The API returns standardized error responses with appropriate HTTP status codes.

### Error Response Format

```json
{
  "code": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "additional_context"
  }
}
```

### Common Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `invalid_request` | Request body cannot be parsed |
| 400 | `validation_error` | Request validation failed |
| 500 | `internal_error` | Internal server error |
| 503 | `service_unavailable` | All providers are unavailable |
| 504 | `timeout` | Request timed out |

### Error Examples

#### Example 4: Invalid Request - Malformed JSON (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS"
    # Missing closing brace
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "invalid_request",
  "message": "Failed to parse request body"
}
```

#### Example 5: Validation Error - Invalid Airport Code (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "JAKARTA",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "validation_error",
  "message": "origin must be a valid 3-letter IATA code, got \"JAKARTA\""
}
```

#### Example 6: Validation Error - Invalid Passenger Count (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 15,
    "class": "economy"
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "validation_error",
  "message": "passengers must be at most 9"
}
```

#### Example 7: Validation Error - Invalid Date Format (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "15-12-2025",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "validation_error",
  "message": "departureDate must be in YYYY-MM-DD format"
}
```

#### Example 8: Validation Error - Invalid Cabin Class (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "premium-economy"
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "validation_error",
  "message": "class must be one of: economy, business, first"
}
```

#### Example 9: Validation Error - Invalid Time Format (400)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy",
    "filters": {
      "departureTimeRange": {
        "start": "6:00 AM",
        "end": "12:00 PM"
      }
    }
  }'
```

**Response: 400 Bad Request**
```json
{
  "code": "validation_error",
  "message": "departureTimeRange.start must be in HH:MM format (24-hour)"
}
```

#### Example 10: Service Unavailable - All Providers Failed (503)

**Request:**
```bash
# This would occur when all provider services are down
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response: 503 Service Unavailable**
```json
{
  "code": "service_unavailable",
  "message": "all providers failed to respond"
}
```

**Note:** This error occurs when:
- All airline provider APIs are down
- Network connectivity issues to all providers
- All providers timeout simultaneously

#### Example 11: Gateway Timeout - Search Exceeded Time Limit (504)

**Request:**
```bash
# This would occur if the search takes longer than the global timeout (5s default)
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response: 504 Gateway Timeout**
```json
{
  "code": "timeout",
  "message": "search timeout exceeded"
}
```

**Note:** This error occurs when:
- Global search timeout is reached (default 5 seconds)
- Multiple providers are slow to respond
- Network latency issues

#### Example 12: Internal Server Error (500)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy"
  }'
```

**Response: 500 Internal Server Error**
```json
{
  "code": "internal_error",
  "message": "an unexpected error occurred"
}
```

**Note:** This error occurs when:
- Unexpected server-side exceptions
- Database connectivity issues
- Memory allocation failures

## Swagger UI

Interactive API documentation is available via Swagger UI.

### Accessing Swagger UI

1. Start the API server:
   ```bash
   go run cmd/api/main.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

3. The Swagger UI provides:
   - Interactive API exploration
   - Request/response schema documentation
   - Try-it-out functionality for testing endpoints
   - Model definitions and examples

### Generating Updated Documentation

If you make changes to API annotations, regenerate the Swagger docs:

```bash
swag init -g internal/api/server.go -o docs --parseDependency --parseInternal
```

## Data Validation Rules

### Airport Codes
- Must be exactly 3 uppercase letters (IATA format)
- Examples: `CGK`, `DPS`, `SUB`, `BDO`

### Dates
- Format: `YYYY-MM-DD`
- Example: `2025-01-15`

### Passengers
- Minimum: 1
- Maximum: 9

### Cabin Class
- Valid values: `economy`, `business`, `first`
- Case-insensitive

### Sort By
- Valid values: `best`, `price`, `duration`, `departure`
- Case-insensitive

### Time Ranges
- Format: `HH:MM` (24-hour format)
- Example: `06:00`, `22:30`

### Price
- Must be non-negative number
- Currency: IDR (Indonesian Rupiah)

## Response Headers

Standard response headers include:
- `Content-Type: application/json`
- `X-Request-ID`: Unique request identifier for tracking

## Best Practices

1. **Always include required fields**: origin, destination, departureDate, passengers
2. **Use proper IATA codes**: Validate airport codes before sending requests
3. **Handle timeouts gracefully**: Implement retry logic with exponential backoff
4. **Cache results appropriately**: Flight data can change frequently
5. **Validate dates**: Ensure departure dates are in the future
6. **Use filters wisely**: Combine filters to narrow down results effectively
7. **Check metadata**: Use `providers_failed` to understand search quality

## Support

For API support or questions:
- Email: support@flightsearch.example.com
- Documentation: http://localhost:8080/swagger/index.html
