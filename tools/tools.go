package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

// TODO The description for each instance should be relative to their repective instance name when running specific instance commands
//
// Define a struct to hold each command's details
type Command struct {
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
	MonitorTag  string `yaml:"monitor_tag"`
}

// Define a struct to hold the configuration from the YAML file for instance_commands and server_commands
type CommandConfig struct {
	InstanceCommands []Command `yaml:"instance_commands"`
	ServerCommands   []Command `yaml:"server_commands"`
}

type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}

// Function to read instance_commands from the YAML file
// Function to read instance_commands from the YAML file
func ReadInstanceCommandsFromYAML(filePath, instanceArg string) ([]Command, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config CommandConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	// Update descriptions relative to the instance name
	for i := range config.InstanceCommands {
		config.InstanceCommands[i].Description = fmt.Sprintf("%s: %s", instanceArg, config.InstanceCommands[i].Description)
	}

	return config.InstanceCommands, nil
}

// Function to read server_commands from the YAML file
func ReadServerCommandsFromYAML(filePath string) ([]Command, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config CommandConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	return config.ServerCommands, nil
}
func ReadCommandsFromYAML(filePath, instanceArg string) ([]Command, error) {
	commands := make([]Command, 0)

	instanceCommands, err := ReadInstanceCommandsFromYAML(filePath, instanceArg)
	if err != nil {
		return nil, err
	}
	commands = append(commands, instanceCommands...)

	serverCommands, err := ReadServerCommandsFromYAML(filePath)
	if err != nil {
		return nil, err
	}
	commands = append(commands, serverCommands...)

	return commands, nil
}

func ExecuteShellCommand(command string) (string, string, error) {
	cmd := exec.Command("sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

// Function to execute commands and encode output to Base64
func ExecuteAndEncodeCommands(commands []Command) ([]string, error) {
	var base64Outputs []string

	for _, cmd := range commands {
		output, _, err := ExecuteShellCommand(cmd.Command)
		if err != nil {
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
	jsonString, err := json.MarshalIndent(data, "", "    ") // Use four spaces for indentation
	if err != nil {
		return err
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	return ioutil.WriteFile(filePath, jsonString, 0644)
}

func ReadJSONFromFile(filePath string) ([]JSONData, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var jsonData []JSONData
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}

func AppendParsedDataToFile(parsedData []JSONData, filePath string) error {
	// Get the existing JSON data from the file (if it exists)
	existingJSONData, err := ReadJSONFromFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading existing JSON data from %s: %s", filePath, err)
	}

	// Append the new JSON data to the existing data
	allJSONData := append(existingJSONData, parsedData...)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(allJSONData, filePath); err != nil {
		return fmt.Errorf("error writing JSON data to %s: %s", filePath, err)
	}

	return nil
}
