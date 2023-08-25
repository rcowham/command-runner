package tools

import (
	"command-runner/schema"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}

// Global variables to store the states
var P4dInstalled = false
var P4dRunning = false

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
		config.P4Commands[i].Description = fmt.Sprintf("[SDP Instance: %s] %s", instanceArg, config.P4Commands[i].Description)
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

func GetSDPInstances(OutputJSONFilePath string, autobotsArg bool) error {

	logrus.Debugf("Finding p4d instances")

	baseDir := "/p4"
	sdpInstanceList := []string{}

	// Read directory /p4
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		logrus.Fatalf("Could not read directory %s: %v", baseDir, err)
	}

	for _, entry := range entries {
		instancePath := filepath.Join(baseDir, entry.Name(), "root", "db.counters")
		if _, err := os.Stat(instancePath); err == nil {
			// If file exists and is readable
			sdpInstanceList = append(sdpInstanceList, entry.Name())
		}
	}

	// Count instances
	instanceCount := len(sdpInstanceList)
	logrus.Debugf("Found %d SDP instances", instanceCount)

	// Loop through each instance and call workSDPInstance function
	for _, instanceArg := range sdpInstanceList {
		handleSDPInstance(OutputJSONFilePath, instanceArg, autobotsArg)
	}
	return nil
}
func handleSDPInstance(OutputJSONFilePath string, instanceArg string, autobotsArg bool) {
	// Pass the obtained instance to HandleP4Commands
	if err := HandleP4Commands(instanceArg, OutputJSONFilePath); err != nil {
		logrus.Fatalf("Error handling P4 commands for instance %s: %v", instanceArg, err)
	}
	logrus.Debugf("Working on SDP instance: %s", instanceArg)

	// If autobotsArg is true, run the HandleAutobotsScripts
	if autobotsArg {
		logrus.Infof("Running autobots...")
		HandleAutobotsScripts(OutputJSONFilePath, instanceArg, autobotsArg)
	}
}
func FindP4D() {
	// Check if p4d is installed
	_, err := exec.LookPath("p4d")
	if err != nil {
		logrus.Debugf("p4d is not installed.")
		P4dInstalled = false
	} else {
		logrus.Debugf("p4d is installed.")
		P4dInstalled = true
	}

	// Check if a p4d process is running
	cmd := exec.Command("pgrep", "-f", "p4d_*")
	output, _ := cmd.CombinedOutput() // Ignoring errors here as we just need to know if there's output or not

	if strings.TrimSpace(string(output)) != "" {
		logrus.Debugf("p4d service is running.")
		P4dRunning = true
	} else {
		logrus.Debugf("p4d service is not running.")
		P4dRunning = false
	}
}
