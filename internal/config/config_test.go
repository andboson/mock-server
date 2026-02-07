package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExpectations(t *testing.T) {
	c := &Config{}
	jsonData := `[
		{
			"method": "GET",
			"path": "/test",
			"status": 200,
			"mock": "response"
		}
	]`

	err := c.ParseExpectations([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, c.Expectations(), 1)
	require.NotNil(t, c.Expectations()[0].Method)
	require.Equal(t, "GET", *c.Expectations()[0].Method)
	require.NotNil(t, c.Expectations()[0].Path)
	require.Equal(t, "/test", *c.Expectations()[0].Path)
}

func TestLoadExpectationsFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// JSON File
	jsonFile := filepath.Join(tempDir, "expectations.json")
	jsonData := `[{"method": "POST", "path": "/json", "status": 201}]`
	err := os.WriteFile(jsonFile, []byte(jsonData), 0o644)
	require.NoError(t, err)

	c := &Config{}
	err = c.LoadExpectationsFromFile(jsonFile)
	require.NoError(t, err)
	require.Len(t, c.Expectations(), 1)
	require.NotNil(t, c.Expectations()[0].Path)
	require.Equal(t, "/json", *c.Expectations()[0].Path)

	// YAML File
	yamlFile := filepath.Join(tempDir, "expectations.yaml")
	yamlData := `- method: PUT
  path: /yaml
  status: 202
`
	err = os.WriteFile(yamlFile, []byte(yamlData), 0o644)
	require.NoError(t, err)

	err = c.LoadExpectationsFromFile(yamlFile) // Append
	require.NoError(t, err)
	require.Len(t, c.Expectations(), 2)
	require.NotNil(t, c.Expectations()[1].Path)
	require.Equal(t, "/yaml", *c.Expectations()[1].Path)
}

func TestNewConfig_Env(t *testing.T) {
	// Setup ENV
	jsonData := `[{"method": "DELETE", "path": "/env", "status": 204}]`
	t.Setenv("EXPECTATIONS_CONFIG_JSON", jsonData)

	c, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Len(t, c.Expectations(), 1)
	require.NotNil(t, c.Expectations()[0].Path)
	require.Equal(t, "/env", *c.Expectations()[0].Path)
}

func TestNewConfig_EnvFile(t *testing.T) {
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "expectations_file.json")
	jsonData := `[{"method": "GET", "path": "/envfile", "status": 200}]`
	err := os.WriteFile(jsonFile, []byte(jsonData), 0o644)
	require.NoError(t, err)

	t.Setenv("EXPECTATIONS_FILE", jsonFile)

	c, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Len(t, c.Expectations(), 1)
	require.NotNil(t, c.Expectations()[0].Path)
	require.Equal(t, "/envfile", *c.Expectations()[0].Path)
}

func TestLoadExpectationsFromTestData(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		c := &Config{}
		err := c.LoadExpectationsFromFile("../testdata/expectations.json")
		require.NoError(t, err)
		require.Len(t, c.Expectations(), 6)

		e := c.Expectations()[0]
		require.NotNil(t, e.Method)
		require.Equal(t, "GET", *e.Method)
		require.NotNil(t, e.Path)
		require.Equal(t, "/api/hello", *e.Path)
		require.Equal(t, 200, e.StatusCode)
	})

	t.Run("yaml", func(t *testing.T) {
		c := &Config{}
		err := c.LoadExpectationsFromFile("../testdata/expectations.yaml")
		require.NoError(t, err)
		require.Len(t, c.Expectations(), 6)

		e := c.Expectations()[0]
		require.NotNil(t, e.Method)
		require.Equal(t, "GET", *e.Method)
		require.NotNil(t, e.Path)
		require.Equal(t, "/api/hello", *e.Path)
		require.Equal(t, 200, e.StatusCode)
	})
}
