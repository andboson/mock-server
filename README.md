# Mock Server

A simple, configurable HTTP mock server written in Go. It enables you to define request expectations and their corresponding responses using configuration files (YAML/JSON) or environment variables. It also provides a web interface to view request history.

## Features

- **Flexible Matching**: Match requests by HTTP Method, Path (Regex supported), and Body (Regex supported).
- **Custom Responses**: Define the Status Code, Headers, and Body for matched requests.
- **Request History**: View a log of received requests, including timestamps, remote addresses, and matching status.
- **Copy cURL**: Easily copy the cURL command for any received request from the history dashboard.
- **Web Interface**: Simple dashboard to view history.
- **Configurable**: Load expectations via JSON/YAML files or Environment Variables.

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

## Running the Server

```bash
# Using go run
SERVER_ADDR_HTTP=:8080 EXPECTATIONS_FILE=./internal/testdata/expectations.json go run cmd/main.go

# Or build and run
go build -o mock-server cmd/main.go
./mock-server
```

Navigate to `http://localhost:8080` (or your configured port) to see the request history dashboard.
