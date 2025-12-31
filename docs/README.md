# Documentation

This directory contains comprehensive API documentation for the Flight Search System.

## Contents

- **API.md** - Complete API reference documentation
- **swagger.json** - OpenAPI 2.0 specification (generated)
- **swagger.yaml** - OpenAPI 2.0 specification in YAML format (generated)
- **docs.go** - Generated Go code for Swagger integration (generated)
- **examples/** - Request/response examples and usage guides
- **postman/** - Postman collection for API testing

## Quick Start

### Viewing Documentation

#### Option 1: Swagger UI (Recommended)

1. Start the API server:
   ```bash
   go run cmd/api/main.go
   ```

2. Open your browser:
   ```
   http://localhost:8080/swagger/index.html
   ```

#### Option 2: Read the Markdown Docs

See [API.md](API.md) for complete documentation with examples.

#### Option 3: Import Postman Collection

Import `postman/Flight-Search-API.postman_collection.json` into Postman for interactive testing.

## Generating Documentation

### Requirements

Install the Swagger CLI tool:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Or use the Makefile:
```bash
make install-swagger
```

### Generate Swagger Docs

```bash
# Using swag directly
swag init -g internal/api/server.go -o docs --parseDependency --parseInternal

# Or using Makefile
make swagger
```

### Validate Documentation

```bash
make swagger-validate
```

## Documentation Structure

```
docs/
├── API.md                          # Main API documentation
├── swagger.json                    # OpenAPI spec (JSON)
├── swagger.yaml                    # OpenAPI spec (YAML)
├── docs.go                         # Generated Go code
├── examples/
│   ├── README.md                   # Examples guide
│   └── request-examples.json       # Sample requests
└── postman/
    └── Flight-Search-API.postman_collection.json
```

## API Endpoints

### Health
- `GET /health` - Service health check

### Flights
- `POST /api/v1/flights/search` - Search for flights

## Key Features Documented

### Request Parameters
- Origin/destination (IATA codes)
- Departure date
- Passenger count
- Cabin class selection
- Advanced filters (price, stops, time ranges, etc.)
- Result sorting options

### Response Structure
- Search criteria echo
- Metadata (execution stats)
- Flight results with complete details
- Standardized error responses

### Filters
- Maximum price
- Maximum stops
- Airline selection
- Departure/arrival time ranges
- Flight duration constraints

### Sorting Options
- Best match
- Price (low to high)
- Duration (shortest first)
- Departure time

## Examples

See [examples/README.md](examples/README.md) for:
- Basic search examples
- Advanced filtering examples
- Validation test cases
- Common airport/airline codes

## Postman Collection

Import the Postman collection for quick testing:

1. Open Postman
2. Click Import
3. Select `docs/postman/Flight-Search-API.postman_collection.json`
4. Collection includes:
   - All endpoint variations
   - Error test cases
   - Environment variables
   - Pre-configured examples

## API Versioning

Current version: **v1**

Base path: `/api/v1`

## Response Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request (validation error) |
| 500 | Internal Server Error |
| 503 | Service Unavailable |
| 504 | Gateway Timeout |

## Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid_request` | 400 | Request body cannot be parsed |
| `validation_error` | 400 | Request validation failed |
| `internal_error` | 500 | Internal server error |
| `service_unavailable` | 503 | All providers unavailable |
| `timeout` | 504 | Request timed out |

## Development Workflow

1. **Make changes** to handler annotations
2. **Regenerate docs**: `make swagger`
3. **Test locally**: `make swagger-serve`
4. **Validate**: Check Swagger UI at `/swagger/index.html`
5. **Commit**: Include generated docs in version control

## CI/CD Integration

The Swagger documentation should be regenerated in your CI pipeline:

```yaml
# Example GitHub Actions
- name: Generate Swagger Docs
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g internal/api/server.go -o docs --parseDependency --parseInternal
    
- name: Validate Swagger
  run: |
    test -f docs/swagger.json || exit 1
```

## Maintenance

### Adding New Endpoints

1. Add Swagger annotations to handler:
   ```go
   // @Summary		Endpoint summary
   // @Description	Detailed description
   // @Tags		tag-name
   // @Accept		json
   // @Produce		json
   // @Param		paramName	body	Model	true	"Description"
   // @Success		200	{object}	ResponseModel
   // @Failure		400	{object}	ErrorDetail
   // @Router		/path [method]
   ```

2. Regenerate documentation:
   ```bash
   make swagger
   ```

3. Test in Swagger UI

### Updating Models

1. Update struct tags with examples:
   ```go
   Field string `json:"field" example:"value" description:"Field description"`
   ```

2. Regenerate docs
3. Verify in Swagger UI

## Support

For documentation issues:
- Check [Swaggo documentation](https://github.com/swaggo/swag)
- Review [OpenAPI Specification](https://swagger.io/specification/)
- See examples in `examples/` directory

## License

Same as the main project (MIT License)
