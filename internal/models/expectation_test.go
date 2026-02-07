package models

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func TestExpectation_Compile(t *testing.T) {
	tests := []struct {
		name        string
		expectation *Expectation
		wantErr     bool
	}{
		{
			name: "valid path regex",
			expectation: &Expectation{
				Path:    strPtr("^/api/v1/users/\\d+$"),
				Request: strPtr(""),
			},
			wantErr: false,
		},
		{
			name: "invalid path regex",
			expectation: &Expectation{
				Path:    strPtr("["),
				Request: strPtr(""),
			},
			wantErr: true,
		},
		{
			name: "valid request regex",
			expectation: &Expectation{
				Path:    strPtr("/test"),
				Request: strPtr(`{"id":\d+}`),
			},
			wantErr: false,
		},
		{
			name: "invalid request regex",
			expectation: &Expectation{
				Path:    strPtr("/test"),
				Request: strPtr(`{"id": [`),
			},
			wantErr: true,
		},
		{
			name: "wildcards",
			expectation: &Expectation{
				Path:    strPtr(""),
				Request: strPtr(""),
			},
			wantErr: false,
		},
		{
			name: "empty strings",
			expectation: &Expectation{
				Path:    strPtr(""),
				Request: strPtr(""),
			},
			wantErr: false,
		},
		{
			name: "nil fields",
			expectation: &Expectation{
				Path:    nil,
				Request: nil,
			},
			wantErr: false,
		},
		{
			name: "asterisk wildcards",
			expectation: &Expectation{
				Path:    strPtr("*"),
				Request: strPtr("*"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expectation.Compile()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExpectation_Match(t *testing.T) {
	tests := []struct {
		name        string
		expectation Expectation
		method      string
		path        string
		body        string
		want        bool
	}{
		// Method matching
		{
			name: "exact method match",
			expectation: Expectation{
				Method: strPtr("POST"),
				Path:   strPtr("/test"),
			},
			method: "POST",
			path:   "/test",
			want:   true,
		},
		{
			name: "case insensitive method match",
			expectation: Expectation{
				Method: strPtr("post"),
				Path:   strPtr("/test"),
			},
			method: "POST",
			path:   "/test",
			want:   true,
		},
		{
			name: "wildcard method match",
			expectation: Expectation{
				Method: strPtr(""),
				Path:   strPtr("/test"),
			},
			method: "DELETE",
			path:   "/test",
			want:   true,
		},
		{
			name: "method mismatch",
			expectation: Expectation{
				Method: strPtr("POST"),
				Path:   strPtr("/test"),
			},
			method: "GET",
			path:   "/test",
			want:   false,
		},

		// Path matching
		{
			name: "exact path match",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("/users"),
			},
			method: "GET",
			path:   "/users",
			want:   true,
		},
		{
			name: "wildcard path match",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("*"),
			},
			method: "GET",
			path:   "/any/path",
			want:   true,
		},
		{
			name: "regex path match",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr(`^/users/\d+$`),
			},
			method: "GET",
			path:   "/users/123",
			want:   true,
		},
		{
			name: "regex path mismatch",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr(`^/users/\d+$`),
			},
			method: "GET",
			path:   "/users/abc",
			want:   false,
		},
		{
			name: "path mismatch",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("/users"),
			},
			method: "GET",
			path:   "/admin",
			want:   false,
		},
		{
			name: "empty path mismatch",
			expectation: Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("/users"),
			},
			method: "GET",
			path:   "",
			want:   false,
		},

		// Body matching
		{
			name: "exact body match",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr(`{"foo":"bar"}`),
			},
			method: "POST",
			path:   "/data",
			body:   `{"foo":"bar"}`,
			want:   true,
		},
		{
			name: "wildcard body match",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr(""),
			},
			method: "POST",
			path:   "/data",
			body:   `{"random":"json"}`,
			want:   true,
		},
		{
			name: "regex body match",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr(`"id":\d+`),
			},
			method: "POST",
			path:   "/data",
			body:   `{"id":123}`,
			want:   true,
		},
		{
			name: "body mismatch",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr(`{"foo":"bar"}`),
			},
			method: "POST",
			path:   "/data",
			body:   `{"foo":"baz"}`,
			want:   false,
		},
		{
			name: "GET with query params match",
			expectation: Expectation{
				Method:  strPtr("GET"),
				Path:    strPtr("/data"),
				Request: strPtr(`foo=bar&baz=qux`),
			},
			method: "GET",
			path:   "/data",
			body:   `foo=bar&baz=qux`,
			want:   true,
		},
		{
			name: "GET with query params different order",
			expectation: Expectation{
				Method:  strPtr("GET"),
				Path:    strPtr("/data"),
				Request: strPtr(`foo=bar&baz=qux`),
			},
			method: "GET",
			path:   "/data",
			body:   `baz=qux&foo=bar`,
			want:   true,
		},
		{
			name: "GET with query params mismatch",
			expectation: Expectation{
				Method:  strPtr("GET"),
				Path:    strPtr("/data"),
				Request: strPtr(`foo=bar`),
			},
			method: "GET",
			path:   "/data",
			body:   `foo=baz`,
			want:   false,
		},
		{
			name: "empty body match expectation empty",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr(""),
			},
			method: "POST",
			path:   "/data",
			body:   "",
			want:   true,
		},
		{
			name: "empty body mismatch expectation not empty",
			expectation: Expectation{
				Method:  strPtr("POST"),
				Path:    strPtr("/data"),
				Request: strPtr("something"),
			},
			method: "POST",
			path:   "/data",
			body:   "",
			want:   false,
		},
		{
			name: "DELETE ignores body",
			expectation: Expectation{
				Method:  strPtr("DELETE"),
				Path:    strPtr("/resource"),
				Request: strPtr(`some body`),
			},
			method: "DELETE",
			path:   "/resource",
			body:   `any body`,
			want:   false,
		},
		{
			name: "PUT matches body",
			expectation: Expectation{
				Method:  strPtr("PUT"),
				Path:    strPtr("/resource"),
				Request: strPtr(`{"updated":true}`),
			},
			method: "PUT",
			path:   "/resource",
			body:   `{"updated":true}`,
			want:   true,
		},
		{
			name: "PUT body mismatch",
			expectation: Expectation{
				Method:  strPtr("PUT"),
				Path:    strPtr("/resource"),
				Request: strPtr(`{"updated":true}`),
			},
			method: "PUT",
			path:   "/resource",
			body:   `{"updated":false}`,
			want:   false,
		},
		{
			name: "PATCH matches body",
			expectation: Expectation{
				Method:  strPtr("PATCH"),
				Path:    strPtr("/resource"),
				Request: strPtr(`{"patch":true}`),
			},
			method: "PATCH",
			path:   "/resource",
			body:   `{"patch":true}`,
			want:   true,
		},
		{
			name: "PATCH body mismatch",
			expectation: Expectation{
				Method:  strPtr("PATCH"),
				Path:    strPtr("/resource"),
				Request: strPtr(`{"patch":true}`),
			},
			method: "PATCH",
			path:   "/resource",
			body:   `{"patch":false}`,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile regex if needed for the test case
			err := tt.expectation.Compile()
			require.NoError(t, err)

			got := tt.expectation.Match(tt.method, tt.path, tt.body)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExpectation_String(t *testing.T) {
	tests := []struct {
		name        string
		expectation Expectation
		expected    string
	}{
		{
			name: "all fields set",
			expectation: Expectation{
				Method:     strPtr("GET"),
				Path:       strPtr("/api/test"),
				Request:    strPtr(`{"id":1}`),
				StatusCode: 200,
			},
			expected: `Expectation(Method=GET, Path=/api/test, Request={"id":1}, StatusCode=200)`,
		},
		{
			name: "nil Request field",
			expectation: Expectation{
				Method:     strPtr("GET"),
				Path:       strPtr("/api/test"),
				Request:    nil,
				StatusCode: 200,
			},
			expected: "Expectation(Method=GET, Path=/api/test, Request=*, StatusCode=200)",
		},
		{
			name: "all nil fields",
			expectation: Expectation{
				Method:     nil,
				Path:       nil,
				Request:    nil,
				StatusCode: 404,
			},
			expected: "Expectation(Method=*, Path=*, Request=*, StatusCode=404)",
		},
		{
			name: "empty string fields",
			expectation: Expectation{
				Method:     strPtr(""),
				Path:       strPtr(""),
				Request:    strPtr(""),
				StatusCode: 201,
			},
			expected: "Expectation(Method=, Path=, Request=, StatusCode=201)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expectation.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestExpectation_CheckMockResponse(t *testing.T) {
	t.Run("no @ prefix", func(t *testing.T) {
		e := &Expectation{
			MockResponse: "simple response",
		}
		err := e.CheckMockResponse()
		require.NoError(t, err)
		require.Equal(t, "simple response", e.MockResponse)
	})

	t.Run("valid file", func(t *testing.T) {
		content := "file content"
		tmpfile, err := os.CreateTemp("", "mock_response_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name()) //nolint:errcheck

		_, err = tmpfile.WriteString(content)
		require.NoError(t, err)
		err = tmpfile.Close()
		require.NoError(t, err)

		originalPath := "@" + tmpfile.Name()
		e := &Expectation{
			MockResponse: originalPath,
		}
		err = e.CheckMockResponse()
		require.NoError(t, err)
		require.Equal(t, content, e.MockResponse)
		require.Equal(t, originalPath, e.FileSourceOriginal)
	})

	t.Run("non-existent file", func(t *testing.T) {
		e := &Expectation{
			MockResponse: "@non_existent_file.txt",
		}
		err := e.CheckMockResponse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "reading mock response file")
	})
}

func TestExpectation_CreateID(t *testing.T) {
	e := &Expectation{}
	require.Equal(t, uuid.Nil, e.ID)

	e.CreateID()
	require.NotEqual(t, uuid.Nil, e.ID)

	// Create another ID and verify it's different
	firstID := e.ID
	e.CreateID()
	require.NotEqual(t, firstID, e.ID)
}

func TestExpectation_IncrementMatchedCount(t *testing.T) {
	e := &Expectation{}
	require.Equal(t, 0, e.MatchedCount)

	e.IncrementMatchedCount()
	require.Equal(t, 1, e.MatchedCount)

	e.IncrementMatchedCount()
	require.Equal(t, 2, e.MatchedCount)

	// Verify it works with initial count
	e2 := &Expectation{MatchedCount: 10}
	e2.IncrementMatchedCount()
	require.Equal(t, 11, e2.MatchedCount)
}
