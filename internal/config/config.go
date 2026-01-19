package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"andboson/mock-server/internal/models"

	"gopkg.in/yaml.v2"
)

const (
	expectationsConfig = "EXPECTATIONS_CONFIG_JSON"
	expectationsFile   = "EXPECTATIONS_FILE"
)

type Config struct {
	expectations []models.Expectation
}

func NewConfig() (*Config, error) {
	c := &Config{
		expectations: make([]models.Expectation, 0),
	}

	expectationsDataFile := os.Getenv(expectationsFile)
	if expectationsDataFile != "" {
		if err := c.LoadExpectationsFromFile(expectationsDataFile); err != nil {
			return nil, fmt.Errorf("loading expectations from file: %w", err)
		}
	}

	expectationsData := os.Getenv(expectationsConfig)
	if expectationsData != "" {
		if err := c.ParseExpectations([]byte(expectationsData)); err != nil {
			return nil, fmt.Errorf("parsing expectations from env: %w", err)
		}
	}

	return c, nil
}

func (c *Config) Expectations() []models.Expectation {
	return c.expectations
}

func (ec *Config) ParseExpectations(data []byte) error {
	var expectations []models.Expectation
	err := json.Unmarshal(data, &expectations)
	if err != nil {
		return fmt.Errorf("unmarshaling expectations: %w", err)
	}

	ec.expectations = append(ec.expectations, expectations...)
	return nil
}

func (ec *Config) LoadExpectationsFromFile(fileName string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("reading expectations file: %w", err)
	}

	var expectations []models.Expectation

	if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
		if err := yaml.Unmarshal(data, &expectations); err != nil {
			return fmt.Errorf("loading expectations from file: %w", err)
		}
	}

	if strings.HasSuffix(fileName, ".json") {
		if err := json.Unmarshal(data, &expectations); err != nil {
			return fmt.Errorf("loading expectations from file: %w", err)
		}
	}

	ec.expectations = append(ec.expectations, expectations...)
	return nil
}
