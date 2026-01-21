package client

// Expectation represents a registered mock expectation.
type Expectation struct {
	ID           string            `json:"id"`
	MatchedCount int               `json:"matched_count"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Request      string            `json:"request"` // Regex for request body matching
	StatusCode   int               `json:"status"`
	Headers      map[string]string `json:"headers"`
	MockResponse string            `json:"mock"` // Response body or @filename
}

// ExpectationCreate represents the payload to create a new expectation.
type ExpectationCreate struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Request      string            `json:"request,omitempty"` // Regex for request body matching
	StatusCode   int               `json:"status,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	MockResponse string            `json:"mock,omitempty"` // Response body or @filename
}

// ExpectationID response when an expectation is created.
type ExpectationID struct {
	ID string `json:"id"`
}

// MatchStatus represents the status of an expectation Match.
type MatchStatus struct {
	Matched      bool `json:"matched"`
	MatchedCount int  `json:"matched_count"`
}
