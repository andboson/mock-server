# Mock Server

A simple, configurable HTTP mock server written in Go. It enables you to define request expectations and their corresponding responses using configuration files (YAML/JSON) or environment variables. It also provides a web interface to view request history.

## Features

- **Flexible Matching**: Match requests by HTTP Method, Path (Regex supported), and Body (Regex supported).
- **Custom Responses**: Define the Status Code, Headers, and Body for matched requests.
- **Request History**: View a log of received requests, including timestamps, remote addresses, and matching status.
- **Copy cURL**: Easily copy the cURL command for any received request from the history dashboard.
- **Web Interface**: Simple dashboard to view history.
- **Configurable**: Load expectations via JSON/YAML files or Environment Variables.
- **REST API**: Manage and check expectations dynamically via a RESTful API.

## Configuration

The server is configured using Environment Variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_ADDR_HTTP` | Address and port to listen on. | `:8081` |
| `EXPECTATIONS_FILE` | Path to a JSON or YAML file containing expectations. | - |
| `EXPECTATIONS_CONFIG_JSON` | JSON string containing expectations (useful for single-line config). | - |

## Docker Compose

You can easily run the mock server using Docker Compose. A `docker-compose.yaml` file is provided in the root directory.

1.  Ensure you have Docker and Docker Compose installed.
2.  (Optional) Edit `internal/testdata/expectations.yaml` or create your own expectations file and update the volume mapping in `docker-compose.yaml`.
3.  Run the server:

    ```bash
    docker compose up
    ```

    The server will start on port `8081`.

### Expectation Format

Each expectation is an object with the following fields:

- `method`: HTTP Method (e.g., "POST", "GET"). Use `*` to match any method.
- `path`: URL path to match. Supports Regex (e.g., `^/api/v1/user/\d+$`) or `*` to match any path.
- `request`: Regex to match against the request body. If empty or `*`, matches any body.
- `status`: HTTP Status Code to return (e.g., 200, 404).
- `headers`: Map of HTTP headers to include in the response.
- `mock`: The response body string. Can start with `@` to load from a file (e.g. `@/path/to/response.json`).

#### Example `expectations.json`

```json
[
  {
    "method": "POST",
    "path": "^/api/login$",
    "request": ".*admin.*",
    "status": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "mock": "{\"token\": \"fake-admin-token\"}"
  },
  {
    "method": "GET",
    "path": "/health",
    "status": 200,
    "mock": "OK"
  }
]
```

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
  

## Running the Server

```bash
# Using go run
SERVER_ADDR_HTTP=:8080 EXPECTATIONS_FILE=./internal/testdata/expectations.json go run cmd/main.go

# Or with docker
docker run -p 8081:8081 -e SERVER_ADDR_HTTP=":8081" -e EXPECTATIONS_FILE="/app/expectations.yaml" -v $(pwd)/internal/testdata/expectations.yaml:/app/expectations.yaml mock-server:latest
```

## API

The server provides a REST API to manage expectations dynamically.

### Endpoints

- `POST /api/expectation`: Add a new expectation.
- `GET /api/expectation/{id}`: Check if an expectation was matched.
- `DELETE /api/expectation/{id}`: Remove an expectation.
- `GET /api/expectations`: Get all registered expectations.

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
	fmt.Printf("Matched: %v\n", status.Matched)
}
```


