# Architecture

Describes the code as of commit `ad2e33d`. For usage, config tables and examples see [README.md](README.md); they are not repeated here.

A single-binary, in-memory HTTP mock server (Go, `net/http` only, no framework). Users register **expectations** (method/path/body pattern -> canned response); every other request is matched against them and recorded in a **history**. A small web UI and a JSON API manage both.

## Package map

| Path | Role |
|------|------|
| `cmd/main.go` | Wiring: config -> store -> templates -> server, signal-based shutdown |
| `internal/config` | Loads expectations from `EXPECTATIONS_FILE` (`.yaml`/`.yml`/`.json`) then `EXPECTATIONS_CONFIG_JSON`, appended in that order |
| `internal/models` | `Expectation` (matching logic, regex compile, `@file` loading), `HistoryItem` / `HistoryEntry` |
| `internal/services/expectations` | `Store`: in-memory expectations + history behind one `sync.RWMutex` |
| `internal/services/server` | HTTP handlers: management API, mock serving, UI pages |
| `internal/templates` | `//go:embed *.tmpl`: `index.tmpl` (history dashboard), `expectations.tmpl` (management UI, jQuery calling the API) |
| `pkg/client` | Public Go client for the API (same root module, tagged `v1.x`; README pins `@v1.4.0`) |
| `openapi.yaml` | Hand-written API spec |
| `internal/testdata` | Sample expectations (JSON/YAML) used by tests, README and docker-compose |

Deps: `google/uuid`, `yaml.v2`, `moul.io/http2curl`, `testify` (tests).

## Startup (`cmd/main.go`)

1. `config.NewConfig()`: file first, then env JSON, appended.
2. `HISTORY_LIMIT` is parsed here (not in `config`): default 1000, `0` = unlimited, invalid -> fatal.
3. `store.AddExpectations` compiles regexes, loads `@file` mocks, assigns fresh UUIDs (IDs are never loaded from config).
4. Templates parsed (fail fast), `server.NewServer(SERVER_ADDR_HTTP)` (default `:8081`).
5. SIGINT/SIGTERM -> `srv.Stop` with a 10s graceful-shutdown timeout. Server has `ReadHeaderTimeout` 10s and `IdleTimeout` 120s; no `ReadTimeout`/`WriteTimeout`.

## Routing (`server.go`)

Go 1.22 `ServeMux` method+path patterns:

| Pattern | Handler |
|---------|---------|
| `POST /api/expectation` | add (rejects `@file` mocks) |
| `GET /api/expectation/{id}` | `{matched, matched_count}` |
| `PUT /api/expectation/{id}` | update |
| `DELETE /api/expectation/{id}` | remove |
| `GET /api/expectations` | list |
| `DELETE /api/expectations` | clear all |
| `GET /api/history` | JSON history, oldest first |
| `DELETE /api/history` | clear history |
| `GET /expectations-ui` | management UI |
| `/` (any method) | `createHTTPHandler` |

The catch-all serves the dashboard only when `RequestURI == "/"` exactly and method is GET (so `/?x=1` is treated as a mock request). Everything else goes to `ServeMocks`. Because `/` is method-less, an API path with an unregistered method falls through to the catch-all: verified that `POST /api/history` returns 404 and is recorded in history like any mock request. So the README's "reserved paths can't be mocked" holds only for the registered method+path pairs above.

## Mock request lifecycle (`mocks_handler.go`)

1. Body capped at 10 MiB (`MaxBytesReader`) -> 413. This cap applies to the mock path only; API JSON bodies are unbounded.
2. Match input: for **GET** the "body" is `r.URL.Query().Encode()` (sorted, encoded); otherwise the raw body.
3. `store.FindMatch(method, path, body)` under RLock: **first match in insertion order wins**; no specificity ranking.
4. If matched, `store.RecordMatch` (write lock) increments `MatchedCount`.
5. A `HistoryItem` is built (`httputil.DumpRequest` + generated cURL via `http2curl`) and stored whether or not it matched.
6. Response: miss -> `404`, empty body. Hit -> expectation headers; if the headers map is empty, the request's `Accept` value becomes `Content-Type`. Status `0` -> `200`. Body is `MockResponse`.

## Matching rules (`models/expectation.go`)

- **Method**: case-insensitive; nil / `""` / `*` = any.
- **Path**: nil / `""` / `*` = any; otherwise exact string equality **or** an **unanchored** regex (`MatchString`). Users must write `^...$` themselves.
- **Request** (body or GET query): nil / `""` / `*` = any. A non-empty pattern never matches an empty body. Otherwise: exact equality, then (GET only) order-insensitive query comparison, then regex.
- Regexes are compiled once in `Compile()` (called by the store on add/update) and held in unexported fields.

## Data model

- `Expectation`: `ID` (uuid, generated), `MatchedCount` (`json:"-"`, so **not** in any JSON output), `Method/Path/Request` (`*string`), `StatusCode`, `ResponseHeaders`, `MockResponse`, `FileSourceOriginal` (internal: the `@path` string when file-backed).
- `HistoryItem` embeds `http.Request` plus original body, mock body, dump, cURL, matched flag, expectation ID, date. `HistoryEntry` is its JSON view (`Entry()`); `pkg/client.HistoryEntry` mirrors it by hand.
- `@file` mocks: `CheckMockResponse` reads the file at add/update time and stores the content in `MockResponse`; `DumpAvailableExpectations` substitutes the `@path` back for display/export. The API rejects `@...` mocks (arbitrary file read) except a PUT that re-sends the expectation's existing `FileSourceOriginal`.

## Store and concurrency (`expectations/store.go`)

- One `RWMutex` guards expectations (`[]*Expectation`) and history (`[]HistoryItem`).
- `UpdateExpectation` swaps the pointer, preserving ID and `MatchedCount`.
- History is capped by dropping from the front when over `historyLimit`.
- `GetHistory` always returns a copy (reversed when `reverse` is true), so callers can iterate without the lock.
- `DumpAvailableExpectations` returns value copies.

## Keeping the API surface in sync

The API shape is described in four places that are updated by hand and have drifted: handlers (`internal/services/server`), `openapi.yaml`, `pkg/client` (+ its README section), and `README.md`. Change all four together. Known drift today:

- `openapi.yaml` has no `PUT /api/expectation/{id}`, though the handler, README and UI use it.
- README and the client's `Expectation.MatchedCount` say `GET /api/expectations` includes `matched_count`, but the field is `json:"-"`, so it is never emitted (the client always reads `0`). Only the server-rendered `expectations.tmpl` card shows it; the UI's own JS reads `/api/expectations` only to populate the edit form.

## Build, test, CI, release

- `make build` (CGO off, `-ldflags -X=main.Version/Revision`; those vars don't exist in `main`, so they are no-ops), `make test` (`-race` + coverage), `make lint` (golangci-lint), `make fmt`, `make audit`.
- `make generate` is dead (no `go:generate` directives exist, `internal/mocks` doesn't exist) but is still called from `make all` and from CI; remove it with the CI steps if touched.
- CI (`ContinuousIntegration.yaml`, PRs only): gofmt check, `make generate`, golangci-lint, `make test` (`go test -race` over all packages except cmd/mocks/generated).
- Release workflow: binaries + changelog + Docker Hub push. Dockerfile is multi-stage -> distroless static, runs as UID 1001.
- Go style conventions: `.github/copilot-instructions.md`.

## Known limitations / issues

- `FindMatch` and `RecordMatch` take the lock separately: if an expectation is updated or removed between them, the increment lands on the stale pointer and is lost.
- No match precedence beyond insertion order; path regexes are unanchored.
- API request bodies have no size limit; no `WriteTimeout`.
- All state is in memory; a restart restores only config-file/env expectations.
- Unknown `EXPECTATIONS_FILE` extension is silently ignored (loads nothing, no error).
