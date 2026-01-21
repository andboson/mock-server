package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"andboson/mock-server/internal/models"
)

// AddExpectationHandler handles the creation of a new expectation via API.
func (h *Server) AddExpectationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.Expectation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.store.AddExpectation(&req); err != nil {
		log.Printf("Failed to add expectation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to add expectation: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return the ID of the created expectation
	if err := json.NewEncoder(w).Encode(map[string]any{
		"id": req.ID,
	}); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// CheckExpectationHandler checks if an expectation was matched.
func (h *Server) CheckExpectationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	exp, err := h.store.GetExpectation(id)
	if err != nil {
		http.Error(w, "Expectation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"matched":       exp.MatchedCount > 0,
		"matched_count": exp.MatchedCount,
	}); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// RemoveExpectationHandler removes an expectation.
func (h *Server) RemoveExpectationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if err := h.store.RemoveExpectation(id); err != nil {
		http.Error(w, "Expectation not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllExpectationsHandler returns all available expectations.
func (h *Server) GetAllExpectationsHandler(w http.ResponseWriter, r *http.Request) {
	exps := h.store.DumpAvailableExpectations()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exps); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
