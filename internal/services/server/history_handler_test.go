package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andboson/mock-server/internal/models"
	"github.com/andboson/mock-server/internal/services/expectations"
	"github.com/andboson/mock-server/internal/templates"

	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) (*Server, *expectations.Store) {
	t.Helper()
	tpls, err := templates.NewTemplates()
	require.NoError(t, err)
	store := expectations.NewStore()

	return NewServer(":0", tpls, store), store
}

func do(srv *Server, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.server.Handler.ServeHTTP(w, req)

	return w
}

func TestServer_History_Empty(t *testing.T) {
	srv, _ := newTestServer(t)
	w := do(srv, http.MethodGet, "/api/history", "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))
	require.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
}

func TestServer_History_PopulatedAndClear(t *testing.T) {
	srv, store := newTestServer(t)
	exp := models.Expectation{Method: strPtr("POST"), Path: strPtr("^/some/path$"), MockResponse: "ok"}
	require.NoError(t, store.AddExpectation(&exp))

	require.Equal(t, http.StatusOK, do(srv, http.MethodPost, "/some/path", `{"a":1}`).Code)
	require.Equal(t, http.StatusNotFound, do(srv, http.MethodGet, "/some/other", "").Code)

	w := do(srv, http.MethodGet, "/api/history", "")
	var entries []models.HistoryEntry
	require.NoError(t, json.NewDecoder(w.Body).Decode(&entries))
	require.Len(t, entries, 2, "API calls must not be recorded")
	require.True(t, entries[0].Matched)
	require.Equal(t, exp.ID.String(), entries[0].ExpectationID)
	require.Equal(t, `{"a":1}`, entries[0].Body)
	require.Equal(t, "/some/path", entries[0].Path)
	require.False(t, entries[1].Matched)
	require.Empty(t, entries[1].ExpectationID)

	require.Equal(t, http.StatusNoContent, do(srv, http.MethodDelete, "/api/history", "").Code)
	w = do(srv, http.MethodGet, "/api/history", "")
	require.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
}

func TestServer_ClearExpectations(t *testing.T) {
	srv, store := newTestServer(t)
	exp := models.Expectation{Path: strPtr("/x"), MockResponse: "ok"}
	require.NoError(t, store.AddExpectation(&exp))
	require.Equal(t, http.StatusOK, do(srv, http.MethodGet, "/x", "").Code)

	require.Equal(t, http.StatusNoContent, do(srv, http.MethodDelete, "/api/expectations", "").Code)
	w := do(srv, http.MethodGet, "/api/expectations", "")
	require.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
	require.Equal(t, http.StatusNotFound, do(srv, http.MethodGet, "/x", "").Code)
}
