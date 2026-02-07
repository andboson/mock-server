package models

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"moul.io/http2curl"
)

type HistoryItem struct {
	http.Request
	BodyOriginal string
	BodyMock     string
	Dump         string
	CurlCommand  string
	MockMatched  bool
	Date         time.Time
}

func (hi *HistoryItem) String() string {
	return fmt.Sprintf("HistoryItem(Date=%s, MockMatched=%t, CurlCommand=%s)", hi.Date.Format(time.RFC3339), hi.MockMatched, hi.CurlCommand)
}

// PrintString returns unescaped html string
func (hi *HistoryItem) PrintString() template.HTML {
	buff := bytes.NewBuffer(nil)
	fmt.Fprintf(
		buff,
		"<pre >request:<code class=\"language-json\">%s</code></pre>",
		hi.Dump,
	)

	if hi.BodyMock != "" {
		fmt.Fprintf(
			buff,
			"<pre>mock response:<code class=\"language-json\">%s</code></pre>",
			hi.BodyMock,
		)
	} else if !hi.MockMatched {
		fmt.Fprintf(
			buff,
			"<pre >mock response:<code class=\"language-json\">%s</code></pre>",
			"request wasn't matched",
		)
	} else {
		fmt.Fprintf(
			buff,
			"<pre >mock response:<code class=\"language-json\">%s</code></pre>",
			"request was matched but mock response is empty",
		)
	}

	return template.HTML(buff.String())
}

func HistoryItemFromHTTPRequest(req http.Request) (*HistoryItem, error) {
	if req.Body == nil {
		req.Body = http.NoBody
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	_ = req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(data))

	curlCommand, err := http2curl.GetCurlCommand(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate curl command: %w", err)
	}

	dump, err := httputil.DumpRequest(&req, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump request: %w", err)
	}

	return &HistoryItem{
		Request:      req,
		BodyOriginal: string(data),
		Dump:         string(dump),
		Date:         time.Now(),
		CurlCommand:  curlCommand.String(),
	}, nil
}

type History struct {
	Values []HistoryItem
}

func (h *History) AddItem(item HistoryItem) {
	h.Values = append(h.Values, item)
}
