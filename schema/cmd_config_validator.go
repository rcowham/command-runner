package schema

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ValidateCmdConfigYAML validates the structure and content of CmdConfig.yaml
func ValidateCmdConfigYAML(filePath string) error {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("Failed to read YAML file: %v", err)
		return err
	}

	// Unmarshal the YAML to the CmdConfigConfig struct
	var config CmdConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		logrus.Errorf("Error parsing YAML: %v", err)
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	// Validate commands and monitor tags
	if err := validateInstanceCommands(config.P4Commands); err != nil {
		logrus.Error(err)
		return err
	}

	if err := validateServerCommands(config.OsCommands); err != nil {
		logrus.Error(err)
		return err
	}

	if err := validateFileParser(config.Files); err != nil {
		logrus.Error(err)
		return err
	}

	// Validate parsing level
	if err := EnsureParsingLevel(config); err != nil {
		logrus.Error(err)
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
			err := fmt.Errorf("missing command for instance command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
		if isEmpty(cmd.MonitorTag) {
			err := fmt.Errorf("missing monitor_tag for instance command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
		if isEmpty(cmd.Description) {
			err := fmt.Errorf("missing description for instance command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
	}
	return nil
}

// Validations for Server Commands
func validateServerCommands(commands []Command) error {
	for _, cmd := range commands {
		if isEmpty(cmd.Command) {
			err := fmt.Errorf("missing command for server command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
		if isEmpty(cmd.MonitorTag) {
			err := fmt.Errorf("missing monitor_tag for server command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
		if isEmpty(cmd.Description) {
			err := fmt.Errorf("missing description for server command: %s", cmd.Description)
			logrus.Error(err)
			return err
		}
	}
	return nil
}

// Validations for Fileparser
func validateFileParser(files []FileConfig) error {
	for _, file := range files {
		if isEmpty(file.MonitorTag) {
			err := fmt.Errorf("missing monitor_tag for file path: %s", file.PathToFile)
			logrus.Error(err)
			return err
		}
		if file.PathToFile == "" {
			err := fmt.Errorf("missing pathtofile for file path: %s", file.PathToFile)
			logrus.Error(err)
			return err
		}
		if file.ParseAll && len(file.Keywords) > 0 {
			logrus.Infof("Warning: For file %s, parseAll is set to true, but keywords are provided. Keywords will be ignored.", file.PathToFile)
		} else if !file.ParseAll && len(file.Keywords) == 0 {
			err := fmt.Errorf("for file %s: parseAll is set to false, but no keywords are provided", file.PathToFile)
			logrus.Error(err)
			return err
		}
		if file.ParsingLevel != "server" && file.ParsingLevel != "instance" {
			err := fmt.Errorf("invalid parsingLevel '%s' for file path: %s. Expecting 'server' or 'instance'", file.ParsingLevel, file.PathToFile)
			logrus.Error(err)
			return err
		}
	}
	return nil
}
func EnsureParsingLevel(config CmdConfig) error {
	for _, file := range config.Files {
		if file.ParsingLevel == "" {
			err := fmt.Errorf("missing parsingLevel for file path: %s", file.PathToFile)
			logrus.Error(err)
			return err
		} else if file.ParsingLevel != "server" && file.ParsingLevel != "instance" {
			err := fmt.Errorf("invalid parsingLevel for file path %s: %s", file.PathToFile, file.ParsingLevel)
			logrus.Error(err)
			return err
		}
	}
	return nil
}
