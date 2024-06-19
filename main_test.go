package main

import (
	"os"
	"testing"
)

func TestSaveConfig(t *testing.T) {
	configPath := "./test_config.json"
	config := &Config{
		APIURL: "http://localhost:8080",
		Model:  "test-model",
	}
	err := saveConfig(configPath, config)
	if err != nil {
		t.Fatalf("saveConfig failed: %s", err)
	}
	os.Remove(configPath)
}
func TestLoadConfig(t *testing.T) {
	configPath := "./test_load_config.json"
	config := &Config{
		APIURL: "http://localhost:8080",
		Model:  "test-model",
	}
	saveConfig(configPath, config)
	loadedConfig, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %s", err)
	}
	if loadedConfig.APIURL != config.APIURL || loadedConfig.Model != config.Model {
		t.Errorf("loadedConfig does not match saved config")
	}
	os.Remove(configPath)
}
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{"single backtick", "`echo hello`", "echo hello"},
		{"triple backticks with lang", "```bash\necho hello```", "echo hello"},
		{"no command", "Just some text", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			command := parseCommand(tc.response)
			if command != tc.expected {
				t.Errorf("parseCommand() = %q, want %q", command, tc.expected)
			}
		})
	}
}
