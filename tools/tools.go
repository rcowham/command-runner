package tools

import (
	"command-runner/schema"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}

// Function to read p4_commands (formarly instance_commands) from the YAML file
func ReadP4CommandsFromYAML(filePath, instanceArg string) ([]schema.Command, error) {
	logrus.Debug("Reading P4 commands from YAML...")
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Error("Failed to read the YAML file:", err)
		return nil, err
	}

	var config schema.CommandConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Error("Failed to unmarshal YAML:", err)
		return nil, err
	}

	// Update descriptions relative to the P4 SDP instance name
	for i := range config.P4Commands {
		config.P4Commands[i].Description = fmt.Sprintf("p4d_%s: %s", instanceArg, config.P4Commands[i].Description)
	}

	logrus.Info("Successfully read P4 commands from YAML.")
	return config.P4Commands, nil
}

// Function to read os_commands (formerly server_commands) from the YAML file
func ReadOsCommandsFromYAML(filePath string) ([]schema.Command, error) {
	logrus.Debug("Reading OS commands from YAML...")
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config schema.CommandConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Error("Failed to unmarshal YAML:", err)
		return nil, err
	}

	logrus.Info("Successfully read OS commands from YAML.")
	return config.OsCommands, nil
}
func ReadCommandsFromYAML(filePath, instanceArg string) ([]schema.Command, error) {
	logrus.Debug("Reading commands from YAML...")
	commands := make([]schema.Command, 0)

	p4Commands, err := ReadP4CommandsFromYAML(filePath, instanceArg)
	if err != nil {
		logrus.Errorf("Error reading P4 commands: %s", err)
		return nil, err
	}
	commands = append(commands, p4Commands...)

	osCommands, err := ReadOsCommandsFromYAML(filePath)
	if err != nil {
		logrus.Errorf("Error reading P4 commands: %s", err)
		return nil, err
	}
	commands = append(commands, osCommands...)

	return commands, nil
}

// Encode to Base64
func EncodeToBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
