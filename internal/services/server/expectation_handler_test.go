package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andboson/mock-server/internal/models"
	"github.com/andboson/mock-server/internal/services/expectations"

	"github.com/stretchr/testify/require"
)

func TestServer_AddExpectationHandler(t *testing.T) {
	t.Run("Valid expectation", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		expJSON := `{"method":"GET","path":"/new","status":200,"mock":"ok"}`
		req := httptest.NewRequest(http.MethodPost, "/api/expectation", bytes.NewBufferString(expJSON))
		w := httptest.NewRecorder()

		srv.AddExpectationHandler(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Check response body for ID
		var respBody map[string]any
		err := json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.Contains(t, respBody, "id")
		require.NotEmpty(t, respBody["id"])

		// Verify expectation was added
		storedExps := store.DumpAvailableExpectations()
		require.Len(t, storedExps, 1)
		require.Equal(t, "/new", *storedExps[0].Path)
		require.Equal(t, respBody["id"], storedExps[0].ID.String())
	})

	t.Run("Invalid method", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		req := httptest.NewRequest(http.MethodGet, "/api/expectation", nil)
		w := httptest.NewRecorder()

		srv.AddExpectationHandler(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		req := httptest.NewRequest(http.MethodPost, "/api/expectation", bytes.NewBufferString(`{invalid`))
		w := httptest.NewRecorder()

		srv.AddExpectationHandler(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestServer_CheckExpectationHandler(t *testing.T) {
	t.Run("Expectation found and matched", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		// Add expectation via handler first to get ID
		expJSON := `{"method":"GET","path":"/test","status":200,"mock":"ok"}`
		req := httptest.NewRequest(http.MethodPost, "/api/expectation", bytes.NewBufferString(expJSON))
		w := httptest.NewRecorder()
		srv.AddExpectationHandler(w, req)

		var addResp map[string]string
		_ = json.NewDecoder(w.Result().Body).Decode(&addResp)
		id := addResp["id"]

		// create http request that matches the expectation
		reqMatch := httptest.NewRequest(http.MethodGet, "/test", nil)
		wMatch := httptest.NewRecorder()
		srv.ServeMocks(wMatch, reqMatch)

		respMatch := wMatch.Result()
		defer func() { _ = respMatch.Body.Close() }()
		require.Equal(t, http.StatusOK, respMatch.StatusCode)

		// Check
		reqCheck := httptest.NewRequest(http.MethodGet, "/api/expectation/"+id, nil)
		reqCheck.SetPathValue("id", id)
		wCheck := httptest.NewRecorder()
		srv.CheckExpectationHandler(wCheck, reqCheck)

		respCheck := wCheck.Result()
		defer func() { _ = respCheck.Body.Close() }()

		require.Equal(t, http.StatusOK, respCheck.StatusCode)

		var checkResp map[string]any
		err := json.NewDecoder(respCheck.Body).Decode(&checkResp)
		require.NoError(t, err)
		require.Equal(t, true, checkResp["matched"])
		// JSON numbers are float64
		require.Equal(t, float64(1), checkResp["matched_count"])
	})

	t.Run("Expectation not found", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		id := "invalid-id"
		req := httptest.NewRequest(http.MethodGet, "/api/expectation/"+id, nil)
		req.SetPathValue("id", id)
		w := httptest.NewRecorder()
		srv.CheckExpectationHandler(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Missing ID", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{
			store: store,
		}

		// If path param is missing, we simulate empty ID if possible or null path value
		req := httptest.NewRequest(http.MethodGet, "/api/expectation/", nil)
		// In Go 1.22 mux, if {id} is missing it likely won't match.
		// But here we test the handler directly. If PathValue returns empty string.
		req.SetPathValue("id", "")
		w := httptest.NewRecorder()
		srv.CheckExpectationHandler(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestServer_GetAllExpectationsHandler(t *testing.T) {
	store := expectations.NewStore()
	srv := &Server{
		store: store,
	}

	// Add some expectations
	exps := []string{
		`{"method":"GET","path":"/1","status":200,"mock":"1"}`,
		`{"method":"POST","path":"/2","status":201,"mock":"2"}`,
	}

	for _, expJSON := range exps {
		req := httptest.NewRequest(http.MethodPost, "/api/expectation", bytes.NewBufferString(expJSON))
		w := httptest.NewRecorder()
		srv.AddExpectationHandler(w, req)
		require.Equal(t, http.StatusCreated, w.Result().StatusCode)
	}

	// Get all
	req := httptest.NewRequest(http.MethodGet, "/api/expectations", nil)
	w := httptest.NewRecorder()

	srv.GetAllExpectationsHandler(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var respExps []models.Expectation
	err := json.NewDecoder(resp.Body).Decode(&respExps)
	require.NoError(t, err)
	require.Len(t, respExps, 2)
}

func putExpectation(srv *Server, id, body string) *http.Response {
	req := httptest.NewRequest(http.MethodPut, "/api/expectation/"+id, bytes.NewBufferString(body))
	req.SetPathValue("id", id)
	w := httptest.NewRecorder()
	srv.UpdateExpectationHandler(w, req)
	return w.Result()
}

func TestServer_FileMockRejectedViaAPI(t *testing.T) {
	t.Run("POST rejects @file", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{store: store}
		req := httptest.NewRequest(http.MethodPost, "/api/expectation", bytes.NewBufferString(`{"mock":"@/etc/hostname"}`))
		w := httptest.NewRecorder()
		srv.AddExpectationHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		require.Empty(t, store.DumpAvailableExpectations())
	})

	t.Run("PUT rejects @file on plain expectation", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{store: store}
		exp := &models.Expectation{MockResponse: "plain"}
		require.NoError(t, store.AddExpectation(exp))
		resp := putExpectation(srv, exp.ID.String(), `{"mock":"@/etc/hostname"}`)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		got, err := store.GetExpectation(exp.ID.String())
		require.NoError(t, err)
		require.Equal(t, "plain", got.MockResponse)
	})

	t.Run("PUT unchanged @file on file-backed expectation", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{store: store}
		path := filepath.Join(t.TempDir(), "r.json")
		require.NoError(t, os.WriteFile(path, []byte("hi"), 0o600))
		exp := &models.Expectation{MockResponse: "@" + path}
		require.NoError(t, store.AddExpectation(exp))
		resp := putExpectation(srv, exp.ID.String(), `{"mock":"@`+path+`"}`)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PUT different @file on file-backed expectation", func(t *testing.T) {
		store := expectations.NewStore()
		srv := &Server{store: store}
		path := filepath.Join(t.TempDir(), "r.json")
		require.NoError(t, os.WriteFile(path, []byte("hi"), 0o600))
		exp := &models.Expectation{MockResponse: "@" + path}
		require.NoError(t, store.AddExpectation(exp))
		resp := putExpectation(srv, exp.ID.String(), `{"mock":"@/etc/hostname"}`)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
