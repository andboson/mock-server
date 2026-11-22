package models

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHistoryItem_PrintString_EscapesHTML(t *testing.T) {
	t.Run("escapes request and mock data", func(t *testing.T) {
		item := HistoryItem{
			Dump:        "GET /<script>alert(1)</script> HTTP/1.1",
			BodyMock:    "<img src=x onerror=alert(1)>",
			MockMatched: true,
		}
		out := string(item.PrintString())
		require.NotContains(t, out, "<script>")
		require.NotContains(t, out, "<img")
		require.Contains(t, out, "&lt;script&gt;")
		require.Contains(t, out, "&lt;img")
	})

	t.Run("keeps constant text readable", func(t *testing.T) {
		item := HistoryItem{Dump: "GET / HTTP/1.1"}
		require.Contains(t, string(item.PrintString()), "request wasn't matched")
	})
}

func TestHistoryItem_Entry(t *testing.T) {
	t.Run("maps fields", func(t *testing.T) {
		item := HistoryItem{
			Request:       http.Request{Method: "POST", URL: &url.URL{Path: "/a", RawQuery: "x=1"}, Header: http.Header{"K": {"v"}}},
			BodyOriginal:  "body",
			BodyMock:      "resp",
			CurlCommand:   "curl",
			MockMatched:   true,
			ExpectationID: "id1",
		}
		e := item.Entry()
		require.Equal(t, "POST", e.Method)
		require.Equal(t, "/a", e.Path)
		require.Equal(t, "x=1", e.Query)
		require.Equal(t, []string{"v"}, e.Headers["K"])
		require.Equal(t, "body", e.Body)
		require.Equal(t, "resp", e.Response)
		require.Equal(t, "curl", e.Curl)
		require.True(t, e.Matched)
		require.Equal(t, "id1", e.ExpectationID)
	})

	t.Run("nil URL and headers", func(t *testing.T) {
		e := (&HistoryItem{}).Entry()
		require.NotNil(t, e.Headers)
		require.Empty(t, e.Path)
	})
}
