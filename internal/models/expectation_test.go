package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpectation_Compile(t *testing.T) {
	tests := []struct {
		name        string
		expectation *Expectation
		wantErr     bool
	}{
		{
			name: "valid path regex",
			expectation: &Expectation{
				Path:    "^/api/v1/users/\\d+$",
				Request: "*",
			},
			wantErr: false,
		},
		{
			name: "invalid path regex",
			expectation: &Expectation{
				Path:    "[",
				Request: "*",
			},
			wantErr: true,
		},
		{
			name: "valid request regex",
			expectation: &Expectation{
				Path:    "/test",
				Request: `{"id":\d+}`,
			},
			wantErr: false,
		},
		{
			name: "invalid request regex",
			expectation: &Expectation{
				Path:    "/test",
				Request: `{"id": [`,
			},
			wantErr: true,
		},
		{
			name: "wildcards",
			expectation: &Expectation{
				Path:    "*",
				Request: "*",
			},
			wantErr: false,
		},
		{
			name: "empty strings",
			expectation: &Expectation{
				Path:    "",
				Request: "",
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
				Method: "POST",
				Path:   "/test",
			},
			method: "POST",
			path:   "/test",
			want:   true,
		},
		{
			name: "case insensitive method match",
			expectation: Expectation{
				Method: "post",
				Path:   "/test",
			},
			method: "POST",
			path:   "/test",
			want:   true,
		},
		{
			name: "wildcard method match",
			expectation: Expectation{
				Method: "*",
				Path:   "/test",
			},
			method: "DELETE",
			path:   "/test",
			want:   true,
		},
		{
			name: "method mismatch",
			expectation: Expectation{
				Method: "POST",
				Path:   "/test",
			},
			method: "GET",
			path:   "/test",
			want:   false,
		},

		// Path matching
		{
			name: "exact path match",
			expectation: Expectation{
				Method: "GET",
				Path:   "/users",
			},
			method: "GET",
			path:   "/users",
			want:   true,
		},
		{
			name: "wildcard path match",
			expectation: Expectation{
				Method: "GET",
				Path:   "*",
			},
			method: "GET",
			path:   "/any/path",
			want:   true,
		},
		{
			name: "regex path match",
			expectation: Expectation{
				Method: "GET",
				Path:   `^/users/\d+$`,
			},
			method: "GET",
			path:   "/users/123",
			want:   true,
		},
		{
			name: "regex path mismatch",
			expectation: Expectation{
				Method: "GET",
				Path:   `^/users/\d+$`,
			},
			method: "GET",
			path:   "/users/abc",
			want:   false,
		},
		{
			name: "path mismatch",
			expectation: Expectation{
				Method: "GET",
				Path:   "/users",
			},
			method: "GET",
			path:   "/admin",
			want:   false,
		},
		{
			name: "empty path mismatch",
			expectation: Expectation{
				Method: "GET",
				Path:   "/users",
			},
			method: "GET",
			path:   "",
			want:   false,
		},

		// Body matching
		{
			name: "exact body match",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: `{"foo":"bar"}`,
			},
			method: "POST",
			path:   "/data",
			body:   `{"foo":"bar"}`,
			want:   true,
		},
		{
			name: "wildcard body match",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: "*",
			},
			method: "POST",
			path:   "/data",
			body:   `{"random":"json"}`,
			want:   true,
		},
		{
			name: "regex body match",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: `"id":\d+`,
			},
			method: "POST",
			path:   "/data",
			body:   `{"id":123}`,
			want:   true,
		},
		{
			name: "body mismatch",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: `{"foo":"bar"}`,
			},
			method: "POST",
			path:   "/data",
			body:   `{"foo":"baz"}`,
			want:   false,
		},
		{
			name: "GET ignores body",
			expectation: Expectation{
				Method:  "GET",
				Path:    "/data",
				Request: `should be ignored`,
			},
			method: "GET",
			path:   "/data",
			body:   `anything`,
			want:   true,
		},
		{
			name: "empty body match expectation empty",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: "",
			},
			method: "POST",
			path:   "/data",
			body:   "",
			want:   true,
		},
		{
			name: "empty body mismatch expectation not empty",
			expectation: Expectation{
				Method:  "POST",
				Path:    "/data",
				Request: "something",
			},
			method: "POST",
			path:   "/data",
			body:   "",
			want:   false,
		},
		{
			name: "DELETE ignores body",
			expectation: Expectation{
				Method:  "DELETE",
				Path:    "/resource",
				Request: `some body`,
			},
			method: "DELETE",
			path:   "/resource",
			body:   `any body`,
			want:   true,
		},
		{
			name: "PUT matches body",
			expectation: Expectation{
				Method:  "PUT",
				Path:    "/resource",
				Request: `{"updated":true}`,
			},
			method: "PUT",
			path:   "/resource",
			body:   `{"updated":true}`,
			want:   true,
		},
		{
			name: "PUT body mismatch",
			expectation: Expectation{
				Method:  "PUT",
				Path:    "/resource",
				Request: `{"updated":true}`,
			},
			method: "PUT",
			path:   "/resource",
			body:   `{"updated":false}`,
			want:   false,
		},
		{
			name: "PATCH matches body",
			expectation: Expectation{
				Method:  "PATCH",
				Path:    "/resource",
				Request: `{"patch":true}`,
			},
			method: "PATCH",
			path:   "/resource",
			body:   `{"patch":true}`,
			want:   true,
		},
		{
			name: "PATCH body mismatch",
			expectation: Expectation{
				Method:  "PATCH",
				Path:    "/resource",
				Request: `{"patch":true}`,
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
	e := Expectation{
		Method:     "GET",
		Path:       "/api/test",
		Request:    "*",
		StatusCode: 200,
	}
	expected := "Expectation(Method=GET, Path=/api/test, Request=*, StatusCode=200)"
	require.Equal(t, expected, e.String())
}
