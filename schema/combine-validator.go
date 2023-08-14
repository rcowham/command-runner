package schema

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ValidateCombineYAML validates the structure and content of combine.yaml
func ValidateCombineYAML(filePath string) error {
	// Read the YAML file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal the YAML to the CombineConfig struct
	var config CombineConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	// Validate commands and monitor tags
	if err := validateInstanceCommands(config.InstanceCommands); err != nil {
		return err
	}

	if err := validateServerCommands(config.ServerCommands); err != nil {
		return err
	}

	if err := validateFileParser(config.Files); err != nil {
		return err
	}

	// Validate parsing level
	if err := EnsureParsingLevel(config); err != nil {
		return err
	}

	return nil
}

// Helper function to check if a string is empty after trimming spaces
func isEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// Validate for missing commands, monitor tags, and descriptions

// Validations for Instance Commands
func validateInstanceCommands(commands []Command) error {
	for _, cmd := range commands {
		if isEmpty(cmd.Command) {
			return fmt.Errorf("missing command for instance command: %s", cmd.Description)
		}
		if isEmpty(cmd.MonitorTag) {
			return fmt.Errorf("missing monitor_tag for instance command: %s", cmd.Description)
		}
		if isEmpty(cmd.Description) {
			return fmt.Errorf("missing description for instance command: %s", cmd.Description)
		}
	}
	return nil
}

// Validations for Server Commands
func validateServerCommands(commands []Command) error {
	for _, cmd := range commands {
		if isEmpty(cmd.Command) {
			return fmt.Errorf("missing command for server command: %s", cmd.Description)
		}
		if isEmpty(cmd.MonitorTag) {
			return fmt.Errorf("missing monitor_tag for server command: %s", cmd.Description)
		}
		if isEmpty(cmd.Description) {
			return fmt.Errorf("missing description for server command: %s", cmd.Description)
		}
	}
	return nil
}

// Validations for Fileparser
func validateFileParser(files []FileConfig) error {
	for _, file := range files {
		if isEmpty(file.MonitorTag) {
			return fmt.Errorf("missing monitor_tag for file path: %s", file.PathToFile)
		}
		if file.PathToFile == "" {
			return fmt.Errorf("missing pathtofile for file path: %s", file.PathToFile)
		}

		// Check for parseAll and keywords conditions
		if file.ParseAll && len(file.Keywords) > 0 {
			logrus.Infof("Warning: For file %s, parseAll is set to true, but keywords are provided. Keywords will be ignored.", file.PathToFile)
		} else if !file.ParseAll && len(file.Keywords) == 0 {
			return fmt.Errorf("for file %s: parseAll is set to false, but no keywords are provided", file.PathToFile)
		}

		// Check for valid parsingLevel
		if file.ParsingLevel != "server" && file.ParsingLevel != "instance" {
			return fmt.Errorf("invalid parsingLevel '%s' for file path: %s. Expecting 'server' or 'instance'", file.ParsingLevel, file.PathToFile)
		}
	}
	return nil
}
func EnsureParsingLevel(config CombineConfig) error {
	for _, file := range config.Files {
		if file.ParsingLevel == "" {
			return fmt.Errorf("missing parsingLevel for file path: %s", file.PathToFile)
		} else if file.ParsingLevel != "server" && file.ParsingLevel != "instance" {
			return fmt.Errorf("invalid parsingLevel for file path %s: %s", file.PathToFile, file.ParsingLevel)
		}
	}

	return nil
}
