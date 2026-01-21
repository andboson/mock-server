package server

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"andboson/mock-server/internal/models"
	"andboson/mock-server/internal/services/expectations"

	"github.com/stretchr/testify/require"
)

type errReader struct {
	err error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func TestServer_ServeMocks(t *testing.T) {
	t.Run("Match found", func(t *testing.T) {
		store := expectations.NewStore()
		exp := models.Expectation{
			Method: "POST",
			Path:   "/api/.*",
			//	Request:    `\{"key":"value"\}`,
			StatusCode: http.StatusAccepted,
			ResponseHeaders: map[string]string{
				"X-Custom-Header": "found",
			},
			MockResponse: `{"result":"success"}`,
		}
		require.NoError(t, store.AddExpectation(&exp))

		srv := &Server{
			store: store,
		}

		body := []byte(`{"key":"value"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/data", bytes.NewReader(body))
		w := httptest.NewRecorder()

		srv.ServeMocks(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusAccepted, resp.StatusCode)
		require.Equal(t, "found", resp.Header.Get("X-Custom-Header"))

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, `{"result":"success"}`, string(respBody))

		// Check history
		history := store.GetHistory(false)
		require.Len(t, history, 1)
		require.True(t, history[0].MockMatched)
		require.Equal(t, `{"result":"success"}`, history[0].BodyMock)
		require.Equal(t, "/api/data", history[0].URL.Path)
	})

	t.Run("Match not found", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
		w := httptest.NewRecorder()

		srv.ServeMocks(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Check history
		history := store.GetHistory(false)
		require.Len(t, history, 1)
		require.False(t, history[0].MockMatched)
		require.Empty(t, history[0].BodyMock)
	})

	t.Run("Match found with default status code", func(t *testing.T) {
		store := expectations.NewStore()
		exp := models.Expectation{
			Method:       "GET",
			Path:         "/ping",
			MockResponse: "pong",
			// StatusCode is 0 by default
		}
		require.NoError(t, store.AddExpectation(&exp))

		srv := &Server{
			store: store,
		}

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		srv.ServeMocks(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "pong", string(respBody))
	})

	t.Run("Error reading body", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		req := httptest.NewRequest(http.MethodPost, "/", &errReader{err: errors.New("read error")})
		w := httptest.NewRecorder()

		srv.ServeMocks(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(body), "Error reading request body")
	})
}
