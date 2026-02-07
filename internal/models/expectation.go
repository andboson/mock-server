package models

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Expectation represents a mock rule containing request matching criteria and the expected response.
type Expectation struct {
	ID           uuid.UUID `json:"id" yaml:"-"`
	MatchedCount int       `json:"matched_count"`

	// Request matching criteria
	Method  *string `json:"method,omitempty" yaml:"method,omitempty"`
	Path    *string `json:"path,omitempty" yaml:"path,omitempty"`
	Request *string `json:"request,omitempty" yaml:"request,omitempty"`

	// Response details
	StatusCode      int               `json:"status" yaml:"status"`
	ResponseHeaders map[string]string `json:"headers" yaml:"headers"`
	MockResponse    string            `json:"mock" yaml:"mock"`

	FileSourceOriginal string `json:"-" yaml:"-"` // internal use: original file source if loaded from file

	// Compiled regex patterns for matching
	pathRegex    *regexp.Regexp
	requestRegex *regexp.Regexp
}

func (e *Expectation) String() string {
	method := "*"
	if e.Method != nil {
		method = *e.Method
	}

	path := "*"
	if e.Path != nil {
		path = *e.Path
	}

	request := "*"
	if e.Request != nil {
		request = *e.Request
	}

	return fmt.Sprintf("Expectation(Method=%s, Path=%s, Request=%s, StatusCode=%d)", method, path, request, e.StatusCode)
}

func (e *Expectation) IncrementMatchedCount() {
	e.MatchedCount++
}

func (e *Expectation) CreateID() {
	e.ID = uuid.New()
}

// CheckMockResponse checks if MockResponse contains @ in it and tries to load the file content.
func (e *Expectation) CheckMockResponse() error {
	if strings.HasPrefix(e.MockResponse, "@") {
		filePath := strings.TrimPrefix(e.MockResponse, "@")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading mock response file: %w", err)
		}

		e.FileSourceOriginal = e.MockResponse
		e.MockResponse = string(data)
	}

	return nil
}

// Compile prepares the regular expressions for the Path and Request fields.
// It should be called after loading the Expectation and before using Match.
func (e *Expectation) Compile() error {
	if e.Path != nil && *e.Path != "" && *e.Path != "*" {
		reg, err := regexp.Compile(*e.Path)
		if err != nil {
			return fmt.Errorf("compiling path regex: %w", err)
		}

		e.pathRegex = reg
	}

	if e.Request != nil && *e.Request != "" && *e.Request != "*" {
		reg, err := regexp.Compile(*e.Request)
		if err != nil {
			return fmt.Errorf("compiling request regex: %w", err)
		}

		e.requestRegex = reg
	}

	return nil
}

// Match checks if the incoming request details match this Expectation.
func (e *Expectation) Match(method, path, body string) bool {
	if !e.matchMethod(method) {
		return false
	}

	if !e.matchPath(path) {
		return false
	}

	if e.Request != nil && !e.matchRequestBody(method, body) {
		return false
	}

	return true
}

func (e *Expectation) matchPath(path string) bool {
	if e.Path == nil || *e.Path == "" || *e.Path == "*" {
		return true
	}

	if *e.Path == path {
		return true
	}

	if e.pathRegex != nil && e.pathRegex.MatchString(path) {
		return true
	}

	return false
}

func (e *Expectation) matchRequestBody(method, body string) bool {
	if e.Request == nil || *e.Request == "" {
		return true
	}

	if body == "" && *e.Request != "" {
		return false
	}

	if *e.Request == body {
		return true
	}

	// For GET requests, we can treat the body as a query string and compare key-value pairs regardless of order
	if method == http.MethodGet {
		reqQuery, err := url.ParseQuery(body)
		if err != nil {
			log.Println("failed to parse request query:", err)

			return false
		}

		expectedQuery, err := url.ParseQuery(*e.Request)
		if err != nil {
			log.Println("failed to parse expectation query:", err)

			return false
		}

		if compareQueries(reqQuery, expectedQuery) {
			return true
		}
	}

	if e.requestRegex != nil && e.requestRegex.MatchString(body) {
		return true
	}

	return false
}

func compareQueries(query url.Values, query2 url.Values) bool {
	if len(query) != len(query2) {
		return false
	}

	for key, values := range query {
		values2, ok := query2[key]
		if !ok {
			return false
		}

		if len(values) != len(values2) {
			return false
		}

		if strings.Join(values, "") != strings.Join(values2, "") {
			return false
		}
	}

	return true
}

func (e *Expectation) matchMethod(method string) bool {
	if e.Method == nil || *e.Method == "" {
		return true
	}

	return strings.EqualFold(*e.Method, method)
}
