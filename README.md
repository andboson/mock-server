# Mock Server

A simple, configurable HTTP mock server written in Go. It enables you to define request expectations and their corresponding responses using configuration files (YAML/JSON) or environment variables. It also provides a web interface to view request history.

## Features

- **Flexible Matching**: Match requests by HTTP Method, Path (Regex supported), and Body (Regex supported). For GET requests, query parameters are automatically encoded and matched against the request pattern.
- **Custom Responses**: Define the Status Code, Headers, and Body for matched requests.
- **Match Tracking**: Track how many times each expectation has been matched via the `matched_count` field.
- **Request History**: View a log of received requests, including timestamps, remote addresses, matching status, and the mock response that was returned.
- **Copy cURL**: Easily copy the cURL command for any received request from the history dashboard.
- **Web Interface**: Simple dashboard to view history.
- **Smart Content-Type**: Automatically uses the request's `Accept` header as `Content-Type` when no response headers are specified.
- **Configurable**: Load expectations via JSON/YAML files or Environment Variables.
- **REST API**: Manage and check expectations dynamically via a RESTful API.

## Configuration

The server is configured using Environment Variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_ADDR_HTTP` | Address and port to listen on. | `:8081` |
| `EXPECTATIONS_FILE` | Path to a JSON or YAML file containing expectations. | - |
| `EXPECTATIONS_CONFIG_JSON` | JSON string containing expectations (useful for single-line config). | - |

### Expectation Format

Each expectation is an object with the following fields:

- `method`: HTTP Method (e.g., "POST", "GET"). Leave empty or omit to match any method.
- `path`: URL path to match. Supports Regex (e.g., `^/api/v1/user/\d+$`). Use `*` or leave empty to match any path.
- `request`: Regex to match against the request body. If empty or `*`, matches any body.
  - For GET requests: Query parameters are automatically URL-encoded (e.g., `foo=bar&baz=qux`) and matched against this pattern. Parameter order is normalized for consistent matching.
- `status`: HTTP Status Code to return (e.g., 200, 404). Defaults to 200 if not specified.
- `headers`: Map of HTTP headers to include in the response.
  - If the `headers` map is empty or omitted entirely, the server automatically uses the request's `Accept` header as the `Content-Type` in the response.
- `mock`: The response body string. Can start with `@` to load from a file (e.g. `@/path/to/response.json`).

#### Example `expectations.yaml`

```yaml
- method: POST
  path: /test/.*
  status: 200
  headers:
    Content-Type: application/json
  mock: '{"foo": "bar"}'
  
- method: POST
  path: /test2/.*
  status: 200
  headers:
    Content-Type: application/json
  mock: "@/app/test_response.json"
```

#### GET Requests and Query Parameters

For GET requests, query parameters are automatically URL-encoded and matched against the `request` field. This allows you to match specific query parameter patterns:

```yaml
- method: GET
  path: /api/search
  request: "q=test&limit=10"
  status: 200
  headers:
    Content-Type: application/json
  mock: '{"results": []}'
```

The above expectation will match requests like:
- `GET /api/search?q=test&limit=10`
- `GET /api/search?limit=10&q=test` (parameter order is normalized)

You can also use regex patterns for flexible matching:

```yaml
- method: GET
  path: /api/users
  request: "id=\\d+"
  status: 200
  mock: '{"user": "data"}'
```

This matches any GET request to `/api/users` with an `id` parameter containing digits.

  

## Running the Server

```bash
# Using go run
SERVER_ADDR_HTTP=:8080 EXPECTATIONS_FILE=./internal/testdata/expectations.json go run cmd/main.go

# Or with docker
docker run -p 8081:8081 -e SERVER_ADDR_HTTP=":8081" -e EXPECTATIONS_FILE="/app/expectations.yaml" -v $(pwd)/internal/testdata/expectations.yaml:/app/expectations.yaml andboson/mock-server:latest
```


## Docker Compose

You can easily run the mock server using Docker Compose. A `docker-compose.yaml` file is provided in the root directory.

1.  Ensure you have Docker and Docker Compose installed.
2.  (Optional) Edit `internal/testdata/expectations.yaml` or create your own expectations file and update the volume mapping in `docker-compose.yaml`.
3.  Run the server:

    ```bash
    docker compose up
    ```

    The server will start on port `8081`.

## API

The server provides a REST API to manage expectations dynamically.

### Endpoints

- `POST /api/expectation`: Add a new expectation.
- `GET /api/expectation/{id}`: Check if an expectation was matched. Returns `{"matched": boolean, "matched_count": integer}` indicating whether the expectation was ever matched and how many times.
- `DELETE /api/expectation/{id}`: Remove an expectation.
- `GET /api/expectations`: Get all registered expectations (includes `matched_count` for each).

### OpenAPI Specification

An OpenAPI 3.0 specification is available in `openapi.yaml`. You can use this file with tools like Swagger UI or Postman to interact with the API.

## Go Client

The project includes a Go client library in `pkg/client` for interacting with the mock server programmatically. This is particularly useful for integration tests where you need to dynamicall set up expectations.

### Installation

```bash
go get andboson/mock-server/pkg/client
```

### Usage Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"andboson/mock-server/pkg/client"
)

func main() {
	// Initialize the client
	c := client.New("http://localhost:8081", nil)
	ctx := context.Background()

	// Create an expectation
	exp := client.ExpectationCreate{
		Method:       "GET",
		Path:         "/api/data",
		StatusCode:   200,
		MockResponse: `{"status": "ok"}`,
	}

	idResp, err := c.CreateExpectation(ctx, exp)
	if err != nil {
		log.Fatalf("Failed to create expectation: %v", err)
	}
	fmt.Printf("Expectation created with ID: %s\n", idResp.ID)

	// Check if expectation was matched
	status, err := c.CheckExpectation(ctx, idResp.ID)
	if err != nil {
		log.Fatalf("Failed to check status: %v", err)
	}
	fmt.Printf("Matched: %v, Match Count: %d\n", status.Matched, status.MatchedCount)
}
```


