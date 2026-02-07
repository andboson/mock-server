package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Client_CreateExpectation_Success(t *testing.T) {
	expectedID := "12345"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/expectation", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req ExpectationCreate
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "GET", req.Method)

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(ExpectationID{ID: expectedID})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	resp, err := client.CreateExpectation(context.Background(), ExpectationCreate{
		Method: "GET",
		Path:   "/test",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedID, resp.ID)
}

func Test_Client_UpdateExpectation_Success(t *testing.T) {
	expectedID := "12345"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/expectation/12345", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req ExpectationCreate
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "POST", req.Method)

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(ExpectationID{ID: expectedID})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	resp, err := client.UpdateExpectation(context.Background(), expectedID, ExpectationCreate{
		Method: "POST",
		Path:   "/updated",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedID, resp.ID)
}

func Test_Client_CheckExpectation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/expectation/123", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(MatchStatus{Matched: true, MatchedCount: 5})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	resp, err := client.CheckExpectation(context.Background(), "123")

	require.NoError(t, err)
	assert.True(t, resp.Matched)
	assert.Equal(t, 5, resp.MatchedCount)
}

func Test_Client_RemoveExpectation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/expectation/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	err := client.RemoveExpectation(context.Background(), "123")

	require.NoError(t, err)
}

func Test_Client_GetExpectations_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/expectations", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode([]Expectation{
			{ID: "1", Method: "GET"},
			{ID: "2", Method: "POST"},
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	resp, err := client.GetExpectations(context.Background())

	require.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "GET", resp[0].Method)
	assert.Equal(t, "POST", resp[1].Method)
}

func Test_Client_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := New(server.URL, nil)
	_, err := client.GetExpectations(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "server returned error status: 400")
}
