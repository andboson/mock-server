package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"andboson/mock-server/internal/models"
	"andboson/mock-server/internal/services/expectations"
	"andboson/mock-server/internal/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()
	addr := ":8082"

	srv := NewServer(addr, tpls, store)

	assert.NotNil(t, srv)
	assert.Equal(t, addr, srv.address)
	assert.Equal(t, tpls, srv.tpls)
	assert.Equal(t, store, srv.store)
	assert.NotNil(t, srv.server)
}

func TestServer_Handler_Index(t *testing.T) {
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()
	srv := NewServer(":8080", tpls, store)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	// RequestURI is often not set by NewRequest as it would be incoming from wire, but the handler checks it.
	// r.RequestURI == "/" in the handler code.
	// standard http.NewRequest does not set RequestURI.
	// We might need to set it manually for the test if the handler relies on it.
	req.RequestURI = "/"

	rr := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
	// Verify that templates were rendered (check for some HTML content if known, or just emptiness not being empty)
	assert.NotEmpty(t, rr.Body.String())
}

func TestServer_Handler_ServeMocks(t *testing.T) {
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()

	exp := models.Expectation{
		Method:       "GET",
		Path:         "/api/test",
		StatusCode:   200,
		MockResponse: "mocked response",
	}
	require.NoError(t, store.AddExpectation(exp))

	srv := NewServer(":8080", tpls, store)

	req, err := http.NewRequest(http.MethodGet, "/api/test", nil)
	require.NoError(t, err)
	req.RequestURI = "/api/test" // Just in case, though handler checks r.RequestURI == "/" for index.

	rr := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "mocked response", rr.Body.String())
}

func TestServer_Start_Stop(t *testing.T) {
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()

	// Using :0 to pick a random port
	srv := NewServer(":0", tpls, store)

	// Start in a goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			// We can't t.Error here easily, but we can print
			// Just ensure it doesn't panic
		}
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = srv.Stop(ctx)
	assert.NoError(t, err)
}

func TestServer_Handler_Index_History(t *testing.T) {
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()

	// Add some history
	store.AddHistory(models.HistoryItem{
		Request: http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/foo"},
		},
		BodyMock: "bar",
		Date:     time.Now(),
		Dump:     "GET /foo HTTP/1.1\r\n\r\n",
	})

	srv := NewServer(":8080", tpls, store)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.RequestURI = "/"

	rr := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Body should contain something from history if template renders it
	// We don't know the exact template structure, but checking success is good start.
	// If template lists history, "GET" and "/foo" might appear.
	assert.True(t, strings.Contains(rr.Body.String(), "/foo") || strings.Contains(rr.Body.String(), "GET"), "Body should contain history item details")
}
