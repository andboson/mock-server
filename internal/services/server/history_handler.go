package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andboson/mock-server/internal/models"
)

// GetHistoryHandler returns received requests as JSON, oldest first.
func (h *Server) GetHistoryHandler(w http.ResponseWriter, _ *http.Request) {
	items := h.store.GetHistory(false)
	entries := make([]models.HistoryEntry, 0, len(items))
	for i := range items {
		entries = append(entries, items[i].Entry())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// ClearHistoryHandler removes all recorded requests.
func (h *Server) ClearHistoryHandler(w http.ResponseWriter, _ *http.Request) {
	h.store.ClearHistory()
	w.WriteHeader(http.StatusNoContent)
}

// ClearExpectationsHandler removes all expectations.
func (h *Server) ClearExpectationsHandler(w http.ResponseWriter, _ *http.Request) {
	h.store.ClearExpectations()
	w.WriteHeader(http.StatusNoContent)
}
