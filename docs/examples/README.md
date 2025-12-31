# API Request Examples

This directory contains example request payloads for testing the Flight Search API.

## Files

- `request-examples.json` - Collection of example request payloads for various search scenarios

## Usage

### Using curl

```bash
# Basic search
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 2
  }'

# Search with filters
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d @docs/examples/comprehensive_search.json
```

### Using Postman

1. Import the request examples from `request-examples.json`
2. Set the base URL to `http://localhost:8080/api/v1`
3. Use the POST method for `/flights/search`
4. Select a request example from the body

### Using Swagger UI

1. Navigate to `http://localhost:8080/swagger/index.html`
2. Click on the `/flights/search` endpoint
3. Click "Try it out"
4. Use one of the examples below or modify as needed

## Example Scenarios

### 1. Basic Search
Minimal required fields for a flight search.

### 2. Search with Cabin Class
Specify preferred cabin class (economy, business, first).

### 3. Search with Price Filter
Limit results to flights under a maximum price.

### 4. Direct Flights Only
Filter for non-stop flights only.

### 5. Search with Time Range
Specify preferred departure time window.

### 6. Search with Duration Filter
Limit flights based on total travel time.

### 7. Search by Airlines
Filter results to specific airline codes.

### 8. Sorted Results
Sort results by price, duration, or departure time.

### 9. Comprehensive Search
Combine multiple filters and sorting options.

### 10. Business Class Search
Search for premium cabin options.

### 11. Family Travel
Search with multiple passengers and family-friendly options.

## Testing Validation

The API validates all inputs. Try these to test error handling:

### Invalid IATA Code
```json
{
  "origin": "JAKARTA",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 2
}
```
Expected: 400 Bad Request - "origin must be a valid 3-letter IATA code"

### Invalid Date Format
```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "15-01-2025",
  "passengers": 2
}
```
Expected: 400 Bad Request - "departureDate must be in YYYY-MM-DD format"

### Too Many Passengers
```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 10
}
```
Expected: 400 Bad Request - "passengers must be at most 9"

### Same Origin and Destination
```json
{
  "origin": "CGK",
  "destination": "CGK",
  "departureDate": "2025-12-15",
  "passengers": 2
}
```
Expected: 400 Bad Request - "origin and destination must be different"

### Invalid Time Format
```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 2,
  "filters": {
    "departureTimeRange": {
      "start": "6:00",
      "end": "22:00"
    }
  }
}
```
Expected: 400 Bad Request - "start time must be in HH:MM format"

## Common IATA Airport Codes (Indonesia)

| Code | Airport | City |
|------|---------|------|
| CGK | Soekarno-Hatta | Jakarta |
| DPS | Ngurah Rai | Denpasar (Bali) |
| SUB | Juanda | Surabaya |
| BDO | Husein Sastranegara | Bandung |
| UPG | Sultan Hasanuddin | Makassar |
| KNO | Kualanamu | Medan |
| JOG | Adisucipto | Yogyakarta |
| SRG | Ahmad Yani | Semarang |
| PLM | Sultan Mahmud Badaruddin II | Palembang |
| BTH | Hang Nadim | Batam |

## Airline Codes

| Code | Airline |
|------|---------|
| GA | Garuda Indonesia |
| JT | Lion Air |
| ID | Batik Air |
| AK | AirAsia |
| QZ | Indonesia AirAsia |
