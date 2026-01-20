package models

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Expectation represents a mock rule containing request matching criteria and the expected response.
type Expectation struct {
	// Request matching criteria
	Method  string `json:"method" yaml:"method"`
	Path    string `json:"path" yaml:"path"`
	Request string `json:"request" yaml:"request"`

	// Response details
	StatusCode      int               `json:"status" yaml:"status"`
	ResponseHeaders map[string]string `json:"headers" yaml:"headers"`
	MockResponse    string            `json:"mock" yaml:"mock"`

	// Compiled regex patterns for matching
	pathRegex    *regexp.Regexp
	requestRegex *regexp.Regexp
}

func (e *Expectation) String() string {
	return fmt.Sprintf("Expectation(Method=%s, Path=%s, Request=%s, StatusCode=%d)", e.Method, e.Path, e.Request, e.StatusCode)
}

// CheckMockResponse checks if MockResponse contains @ in it and tries to load the file content.
func (e *Expectation) CheckMockResponse() error {
	if strings.HasPrefix(e.MockResponse, "@") {
		filePath := strings.TrimPrefix(e.MockResponse, "@")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading mock response file: %w", err)
		}

		e.MockResponse = string(data)
	}

	return nil
}

// Compile prepares the regular expressions for the Path and Request fields.
// It should be called after loading the Expectation and before using Match.
func (e *Expectation) Compile() error {
	if e.Path != "" && e.Path != "*" {
		reg, err := regexp.Compile(e.Path)
		if err != nil {
			return fmt.Errorf("compiling path regex: %w", err)
		}

		e.pathRegex = reg
	}

	if e.Request != "" && e.Request != "*" {
		reg, err := regexp.Compile(e.Request)
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

	if e.Method != "POST" && e.Method != "PATCH" && e.Method != "PUT" {
		return true
	}

	if e.Request != "" && !e.matchRequestBody(body) {
		return false
	}

	return true
}

func (e *Expectation) matchPath(path string) bool {
	if e.Path == "*" {
		return true
	}

	if path == "" && e.Path != "" {
		return false
	}

	if e.Path == path {
		return true
	}

	if e.pathRegex != nil && e.pathRegex.MatchString(path) {
		return true
	}

	return false
}

func (e *Expectation) matchRequestBody(body string) bool {
	if e.Request == "*" || e.Method == "GET" {
		return true
	}

	if body == "" && e.Request != "" {
		return false
	}

	if e.Request == body {
		return true
	}

	if e.requestRegex != nil && e.requestRegex.MatchString(body) {
		return true
	}

	return false
}

func (e *Expectation) matchMethod(method string) bool {
	if e.Method == "*" || method == "" {
		return true
	}

	return strings.EqualFold(e.Method, method)
}
