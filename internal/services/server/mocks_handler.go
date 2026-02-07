package server

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"andboson/mock-server/internal/models"
)

// ServeHTTP handles the incoming HTTP request.
func (h *Server) ServeMocks(w http.ResponseWriter, r *http.Request) {
	// Read the body to match against expectations and to create history
	if r.Body == nil {
		r.Body = http.NoBody
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	_ = r.Body.Close()

	// Restore the body for HistoryItemFromHTTPRequest
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	bodyStr := string(bodyBytes)

	if r.Method == http.MethodGet {
		bodyStr = r.URL.Query().Encode()
	}

	// Attempt to match
	exp, found := h.store.FindMatch(r.Method, r.URL.Path, bodyStr)

	// Create history item
	// We pass a dereferenced request (copy), but since we restored the body, it can be read again.
	// Note: HistoryItemFromHTTPRequest will also read and close the body of the copy.
	if r.URL.Host == "" {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
	}

	histItem, err := models.HistoryItemFromHTTPRequest(*r)
	if err != nil {
		log.Printf("Failed to create history item: %v", err)
	} else {
		histItem.MockMatched = found
		if found {
			exp.IncrementMatchedCount()
			histItem.BodyMock = exp.MockResponse
		}
		h.store.AddHistory(*histItem)
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	// Write response headers
	for k, v := range exp.ResponseHeaders {
		w.Header().Set(k, v)
	}

	// If no Content-Type header is set, use the Accept header from the request
	if len(exp.ResponseHeaders) == 0 {
		accept := r.Header.Get("Accept")
		w.Header().Set("Content-Type", accept)
	}

	// Write status code
	statusCode := exp.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	// Write response body
	if _, err := w.Write([]byte(exp.MockResponse)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
