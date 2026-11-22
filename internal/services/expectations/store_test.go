package expectations

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/andboson/mock-server/internal/models"

	"github.com/stretchr/testify/require"
)

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

func TestNewStore(t *testing.T) {
	s := NewStore()
	require.NotNil(t, s)
	require.Empty(t, s.DumpAvailableExpectations())
	require.Empty(t, s.GetHistory(false))
}

func TestStore_AddExpectation(t *testing.T) {
	tests := []struct {
		name        string
		expectation models.Expectation
		wantErr     bool
	}{
		{
			name: "Valid simple expectation",
			expectation: models.Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("/api/test"),
			},
			wantErr: false,
		},
		{
			name: "Valid regex expectation",
			expectation: models.Expectation{
				Method: strPtr("POST"),
				Path:   strPtr("/api/users/\\d+"),
			},
			wantErr: false,
		},
		{
			name: "Invalid path regex",
			expectation: models.Expectation{
				Method: strPtr("GET"),
				Path:   strPtr("/api/["),
			},
			wantErr: true,
		},
		{
			name: "Invalid mock response file",
			expectation: models.Expectation{
				Method:       strPtr("GET"),
				Path:         strPtr("/api/file"),
				MockResponse: "@nonexistent_file.json",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore()
			err := s.AddExpectation(&tt.expectation)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				exps := s.DumpAvailableExpectations()
				require.Len(t, exps, 1)
				require.Equal(t, tt.expectation.Path, exps[0].Path)
			}
		})
	}
}

func TestStore_AddExpectation_FileLoading(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "response.json")
	fileContent := `{"message": "hello from file"}`
	err := os.WriteFile(tmpFile, []byte(fileContent), 0o644)
	require.NoError(t, err)

	s := NewStore()
	exp := models.Expectation{
		Method:       strPtr("GET"),
		Path:         strPtr("/api/file"),
		MockResponse: "@" + tmpFile,
	}

	err = s.AddExpectation(&exp)
	require.NoError(t, err)

	exps := s.DumpAvailableExpectations()
	require.Len(t, exps, 1)
	require.Equal(t, "@"+tmpFile, exps[0].MockResponse)
}

func TestStore_AddExpectations(t *testing.T) {
	s := NewStore()
	expectations := []models.Expectation{
		{Method: strPtr("GET"), Path: strPtr("/a")},
		{Method: strPtr("POST"), Path: strPtr("/b")},
	}

	err := s.AddExpectations(expectations)
	require.NoError(t, err)

	exps := s.DumpAvailableExpectations()
	require.Len(t, exps, 2)
}

func TestStore_AddExpectations_Error(t *testing.T) {
	s := NewStore()
	expectations := []models.Expectation{
		{Method: strPtr("GET"), Path: strPtr("/a")},
		{Method: strPtr("POST"), Path: strPtr("[")}, // Invalid regex
	}

	err := s.AddExpectations(expectations)
	require.Error(t, err)
}

func TestStore_FindMatch(t *testing.T) {
	s := NewStore()
	exp1 := models.Expectation{Method: strPtr("GET"), Path: strPtr("/api/a"), MockResponse: "A"}
	exp2 := models.Expectation{Method: strPtr("POST"), Path: strPtr("/api/b"), MockResponse: "B"}
	exp3 := models.Expectation{Method: strPtr("PUT"), Path: strPtr("/api/c/\\d+"), MockResponse: "C"}

	require.NoError(t, s.AddExpectation(&exp1))
	require.NoError(t, s.AddExpectation(&exp2))
	require.NoError(t, s.AddExpectation(&exp3))

	tests := []struct {
		name      string
		method    string
		path      string
		body      string
		wantMatch bool
		wantMock  string
	}{
		{
			name:      "Match existing GET",
			method:    "GET",
			path:      "/api/a",
			wantMatch: true,
			wantMock:  "A",
		},
		{
			name:      "Match existing POST",
			method:    "POST",
			path:      "/api/b",
			wantMatch: true,
			wantMock:  "B",
		},
		{
			name:      "Match regex path",
			method:    "PUT",
			path:      "/api/c/123",
			wantMatch: true,
			wantMock:  "C",
		},
		{
			name:      "No match method",
			method:    "DELETE",
			path:      "/api/a",
			wantMatch: false,
		},
		{
			name:      "No match path",
			method:    "GET",
			path:      "/api/x",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := s.FindMatch(tt.method, tt.path, tt.body)
			if tt.wantMatch {
				require.True(t, found)
				require.Equal(t, tt.wantMock, got.MockResponse)
			} else {
				require.False(t, found)
			}
		})
	}
}

func TestStore_RemoveExpectation(t *testing.T) {
	s := NewStore()
	exp := &models.Expectation{
		Method: strPtr("GET"),
		Path:   strPtr("/test"),
	}
	require.NoError(t, s.AddExpectation(exp))

	// Get ID
	storedExp, err := s.GetExpectation(exp.ID.String())
	require.NoError(t, err)
	require.NotNil(t, storedExp)

	// Remove existing
	err = s.RemoveExpectation(exp.ID.String())
	require.NoError(t, err)

	// Check it is gone
	_, err = s.GetExpectation(exp.ID.String())
	require.Error(t, err)

	// Remove non-existent
	err = s.RemoveExpectation("non-existent-id")
	require.Error(t, err)
}

func TestStore_History(t *testing.T) {
	s := NewStore()

	h1 := models.HistoryItem{Date: time.Now().Add(-2 * time.Hour), BodyOriginal: "req1"}
	h2 := models.HistoryItem{Date: time.Now().Add(-1 * time.Hour), BodyOriginal: "req2"}

	s.AddHistory(h1)
	s.AddHistory(h2)

	// Test GetHistory without reverse
	history := s.GetHistory(false)
	require.Len(t, history, 2)
	require.Equal(t, "req1", history[0].BodyOriginal)
	require.Equal(t, "req2", history[1].BodyOriginal)

	// Test GetHistory with reverse
	historyRev := s.GetHistory(true)
	require.Len(t, historyRev, 2)
	require.Equal(t, "req2", historyRev[0].BodyOriginal)
	require.Equal(t, "req1", historyRev[1].BodyOriginal)
}

func TestStore_Concurrency(t *testing.T) {
	s := NewStore()
	exp := models.Expectation{Method: strPtr("GET"), Path: strPtr("/test")}
	require.NoError(t, s.AddExpectation(&exp))

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			s.AddHistory(models.HistoryItem{})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			s.FindMatch("GET", "/test", "")
		}
		done <- true
	}()

	<-done
	<-done

	require.Len(t, s.GetHistory(false), 100)
}

func TestStore_RecordMatch_Concurrent(t *testing.T) {
	s := NewStore()
	e := &models.Expectation{Method: strPtr("GET"), Path: strPtr("/x"), StatusCode: 200}
	require.NoError(t, s.AddExpectation(e))
	id := e.ID.String()

	done := make(chan struct{})
	var readers sync.WaitGroup
	readers.Add(1)
	go func() {
		defer readers.Done()
		for {
			select {
			case <-done:
				return
			default:
				_ = s.DumpAvailableExpectations()
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				s.RecordMatch(e)
			}
		}()
	}
	wg.Wait()
	close(done)
	readers.Wait()

	n, err := s.MatchedCount(id)
	require.NoError(t, err)
	require.Equal(t, 1000, n)
}

func TestStore_HistoryLimit(t *testing.T) {
	add := func(s *Store) {
		for _, b := range []string{"1", "2", "3", "4", "5"} {
			s.AddHistory(models.HistoryItem{BodyOriginal: b})
		}
	}

	s := NewStore()
	s.SetHistoryLimit(3)
	add(s)
	h := s.GetHistory(false)
	require.Len(t, h, 3)
	require.Equal(t, "3", h[0].BodyOriginal)
	require.Equal(t, "4", h[1].BodyOriginal)
	require.Equal(t, "5", h[2].BodyOriginal)

	s = NewStore()
	s.SetHistoryLimit(0)
	add(s)
	require.Len(t, s.GetHistory(false), 5)
}

func TestStore_ClearHistory(t *testing.T) {
	s := NewStore()
	s.AddHistory(models.HistoryItem{BodyOriginal: "1"})
	s.AddHistory(models.HistoryItem{BodyOriginal: "2"})
	s.ClearHistory()
	require.Empty(t, s.GetHistory(false))

	s.SetHistoryLimit(1)
	s.AddHistory(models.HistoryItem{BodyOriginal: "3"})
	s.AddHistory(models.HistoryItem{BodyOriginal: "4"})
	require.Len(t, s.GetHistory(false), 1)
}

func TestStore_ClearExpectations(t *testing.T) {
	s := NewStore()
	p := "/x"
	require.NoError(t, s.AddExpectation(&models.Expectation{Path: &p}))
	require.NoError(t, s.AddExpectation(&models.Expectation{Path: &p}))
	_, ok := s.FindMatch("GET", "/x", "")
	require.True(t, ok)

	s.ClearExpectations()
	require.Empty(t, s.DumpAvailableExpectations())
	_, ok = s.FindMatch("GET", "/x", "")
	require.False(t, ok)
}
