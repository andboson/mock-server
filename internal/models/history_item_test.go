package models

import (
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
