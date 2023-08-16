package tools

import (
	"bytes"
	"command-runner/schema"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}

// Function to read instance_commands from the YAML file
func ReadInstanceCommandsFromYAML(filePath, instanceArg string) ([]schema.Command, error) {
	logrus.Debug("Reading instance commands from YAML...")
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

	// Update descriptions relative to the instance name
	for i := range config.InstanceCommands {
		config.InstanceCommands[i].Description = fmt.Sprintf("p4d_%s: %s", instanceArg, config.InstanceCommands[i].Description)
	}

	logrus.Info("Successfully read instance commands from YAML.")
	return config.InstanceCommands, nil
}

// Function to read server_commands from the YAML file
func ReadServerCommandsFromYAML(filePath string) ([]schema.Command, error) {
	logrus.Debug("Reading server commands from YAML...")
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config schema.CommandConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Error("Failed to unmarshal YAML:", err)
		return nil, err
	}

	logrus.Info("Successfully read server commands from YAML.")
	return config.ServerCommands, nil
}
func ReadCommandsFromYAML(filePath, instanceArg string) ([]schema.Command, error) {
	logrus.Debug("Reading commands from YAML...")
	commands := make([]schema.Command, 0)

	instanceCommands, err := ReadInstanceCommandsFromYAML(filePath, instanceArg)
	if err != nil {
		logrus.Errorf("Error reading instance commands: %s", err)
		return nil, err
	}
	commands = append(commands, instanceCommands...)

	serverCommands, err := ReadServerCommandsFromYAML(filePath)
	if err != nil {
		logrus.Errorf("Error reading server commands: %s", err)
		return nil, err
	}
	commands = append(commands, serverCommands...)

	return commands, nil
}

func ExecuteShellCommand(command string) (string, string, error) {
	logrus.Debugf("Executing shell command: %s", command)
	cmd := exec.Command("sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to execute command %s: %s", command, err)
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

// Function to execute commands and encode output to Base64
func ExecuteAndEncodeCommands(commands []schema.Command) ([]string, error) {
	var base64Outputs []string

	for _, cmd := range commands {
		logrus.Debugf("Encoding command: %s", cmd.Command)
		output, _, err := ExecuteShellCommand(cmd.Command)
		if err != nil {
			logrus.Errorf("Error executing and encoding command %s: %s", cmd.Command, err)
			return nil, err
		}
		base64Output := EncodeToBase64(output)
		base64Outputs = append(base64Outputs, base64Output)
	}

	return base64Outputs, nil
}

func EncodeToBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// Function to write JSON data to a file with indentation for human-readability
func WriteJSONToFile(data []JSONData, filePath string) error {
	logrus.Debugf("Writing JSON data to file: %s", filePath)
	jsonString, err := json.MarshalIndent(data, "", "    ") // Use four spaces for indentation
	if err != nil {
		logrus.Errorf("Failed to marshal JSON data: %s", err)
		return err
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(filePath)
		if err != nil {
			logrus.Errorf("Failed to create file %s: %s", filePath, err)
			return err
		}
		defer file.Close()
	}

	return ioutil.WriteFile(filePath, jsonString, 0644)
}

func ReadJSONFromFile(filePath string) ([]JSONData, error) {
	logrus.Debugf("Reading JSON data from file: %s", filePath)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("Failed to open file %s: %s", filePath, err)
		return nil, err
	}
	defer jsonFile.Close()

	var jsonData []JSONData
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&jsonData); err != nil {
		logrus.Errorf("Failed to decode JSON data from file %s: %s", filePath, err)
		return nil, err
	}

	return jsonData, nil
}

func AppendParsedDataToFile(parsedData []JSONData, filePath string) error {
	logrus.Debugf("Appending parsed data to file: %s", filePath)
	// Get the existing JSON data from the file (if it exists)
	existingJSONData, err := ReadJSONFromFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Error reading existing JSON data from %s: %s", filePath, err)
		return fmt.Errorf("error reading existing JSON data from %s: %s", filePath, err)
	}

	// Append the new JSON data to the existing data
	allJSONData := append(existingJSONData, parsedData...)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(allJSONData, filePath); err != nil {
		logrus.Errorf("Error appending parsed data to file %s: %s", filePath, err)
		return fmt.Errorf("error writing JSON data to %s: %s", filePath, err)
	}

	return nil
}
