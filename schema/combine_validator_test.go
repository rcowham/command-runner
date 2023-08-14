package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectFileRead(t *testing.T) {
	testFiles := []string{"valid.yaml", "invalid_missing_command.yaml", "missing_monitor_tag.yaml", "invalid_yaml.yaml", "invalid_missing_parsing_level.yaml", "invalid_parsing_level.yaml", "missing_description.yaml"}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			_, err := ioutil.ReadFile(filepath.Join("testfiles", filename))
			if err != nil {
				t.Fatalf("Failed to read %s: %v", filename, err)
			} else {
				fmt.Printf("Successfully read file %s\n", filename)
			}
		})
	}
}

func TestValidateCombineYAMLFromFile(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{
			name:     "Valid YAML",
			filepath: filepath.Join("testfiles", "valid.yaml"),
			wantErr:  false,
		},
		{
			name:     "Invalid YAML - missing command",
			filepath: filepath.Join("testfiles", "invalid_missing_command.yaml"),
			wantErr:  true,
		},
		{
			name:     "Invalid YAML - missing monitor tag",
			filepath: filepath.Join("testfiles", "missing_monitor_tag.yaml"),
			wantErr:  true,
		},
		{
			name:     "Invalid YAML - invalid formatting",
			filepath: filepath.Join("testfiles", "invalid_yaml.yaml"),
			wantErr:  true,
		},
		{
			name:     "Invalid YAML - missing parsing level",
			filepath: filepath.Join("testfiles", "invalid_missing_parsing_level.yaml"),
			wantErr:  true,
		},
		{
			name:     "Invalid YAML - invalid parsing level",
			filepath: filepath.Join("testfiles", "invalid_parsing_level.yaml"),
			wantErr:  true,
		},
		{
			name:     "Invalid YAML - missing description",
			filepath: filepath.Join("testfiles", "missing_description.yaml"),
			wantErr:  true,
		},
		{
			name:     "Instance Level - parseAll true, no keywords",
			filepath: filepath.Join("testfiles", "parseAll_true_no_keywords.yaml"),
			wantErr:  false,
		},
		{
			name:     "Instance Level - parseAll false, with keywords",
			filepath: filepath.Join("testfiles", "parseAll_false_with_keywords.yaml"),
			wantErr:  false,
		},
		{
			name:     "Instance Level - parseAll false, with no keywords",
			filepath: filepath.Join("testfiles", "parseAll_false_with_no_keywords.yaml"),
			wantErr:  true,
		},
		{
			name:     "Server/Instance Level mismatch - parsingLevel not 'server' or 'instance'",
			filepath: filepath.Join("testfiles", "parseLevel_not_server_or_instance.yaml"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.name)
			// Validate the content
			err := ValidateCombineYAML(tt.filepath)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}
		})
	}
}
