package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config represents the schema for the application settings.
// Using 'yaml' tags allows the parser to map file keys to struct fields.
type Config struct {
	Settings struct {
		Interval int      `yaml:"interval"` // In seconds
		Targets  []string `yaml:"targets"`
	} `yaml:"settings"`
}

// LoadConfig reads and decodes the YAML file from the given path.
// It returns an error if the file is missing or the format is invalid, 
// adhering to the "fail-fast" principle.
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}