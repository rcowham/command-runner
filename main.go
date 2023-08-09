/*
TODO
*/
package main

import (
	"command-runner/tools"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CommandsYAMLPath   string `yaml:"commands_yaml"`
	FileParserYAMLPath string `yaml:"fileparser_yaml"`
	OutputJSONPath     string `yaml:"output_json_path"`
}

var (
	outputJSONFilePath     string
	yamlCommandsFilePath   string
	yamlFileparserFilePath string
	cloudProvider          string
	configFilePath         string
)
var debug bool

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&outputJSONFilePath, "output", "out.json", "Path to the output JSON file")

	flag.StringVar(&cloudProvider, "cloud", "", "Cloud provider (aws, gcp, or azure)")

	flag.StringVar(&configFilePath, "config", "", "Path to the config file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	// Setup the logrus level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

/*
// Function to get the absolute path for a file or directory

	func getAbsolutePath(baseDir, relPath string) (string, error) {
		absPath := filepath.Join(baseDir, relPath)
		return absPath, nil
	}
*/
func main() {
	var instanceArg string
	var serverArg bool

	flag.StringVar(&instanceArg, "instance", "", "Instance argument for the command-runner")
	flag.BoolVar(&serverArg, "server", false, "Server argument for the command-runner")

	flag.Parse()

	// If config file path is not provided, use the default
	if configFilePath == "" {
		exePath, err := os.Executable()
		if err != nil {
			logrus.Errorf("Error getting executable path: %v", err)
			os.Exit(1)
		}

		configFilePath = filepath.Join(filepath.Dir(exePath), "configs", "config.yaml")
	}

	// Read the config file
	configFile, err := os.Open(configFilePath)
	if err != nil {
		logrus.Errorf("Error opening config file: %v", err)
		os.Exit(1)
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		logrus.Errorf("Error decoding config file: %v", err)
		os.Exit(1)
	}

	// Retrieve the paths from the config struct
	configsDir := filepath.Dir(configFilePath)
	yamlCommandsFilePath = filepath.Join(configsDir, config.CommandsYAMLPath)
	yamlFileparserFilePath = filepath.Join(configsDir, config.FileParserYAMLPath)
	if outputJSONFilePath == "" {
		outputJSONFilePath = filepath.Join(config.OutputJSONPath)
	}

	// Get absolute path for "configs/commands.yaml"
	if err != nil {
		logrus.Errorf("Error getting absolute path for commands.yaml: %v", err)
		os.Exit(1)
	}

	// If -cloud is provided, check if it's a valid cloud provider
	if cloudProvider != "" {
		switch cloudProvider {
		case "aws":
			//Logic to handle AWS-related functionality
			err := tools.GetAWSInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				logrus.Errorf("Error getting AWS instance identity info: %v", err)
				os.Exit(1)
			}
		case "gcp":
			//Logic to handle GCP-related functionality
			err := tools.GetGCPInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				logrus.Errorf("Error getting GCP instance identity info: %v", err)
				os.Exit(1)
			}
		case "azure":
			// Add logic to handle Azure-related functionality
		default:
			logrus.Errorf("Error: Invalid cloud provider. Please specify aws, gcp, or azure.")
			os.Exit(1)
		}
	}

	// Check if the server argument is provided
	if serverArg {
		//		fmt.Println("Server ARG passed")
		// Execute and encode server commands
		serverCommands, err := tools.ReadServerCommandsFromYAML(yamlCommandsFilePath)
		if err != nil {
			logrus.Errorf("Error reading server commands from YAML: %v", err)
			os.Exit(1)
		}

		base64ServerOutputs, err := tools.ExecuteAndEncodeCommands(serverCommands)
		if err != nil {
			logrus.Errorf("Error executing server commands: %v", err)
			os.Exit(1)
		}

		// Create JSON data for server commands
		var serverJSONData []tools.JSONData
		for i, cmd := range serverCommands {
			serverJSONData = append(serverJSONData, tools.JSONData{
				Command:     cmd.Command,
				Description: cmd.Description,
				Output:      base64ServerOutputs[i],
				MonitorTag:  cmd.MonitorTag,
			})
		}

		// Get the existing JSON data from the file (if it exists)
		existingJSONData, err := tools.ReadJSONFromFile(outputJSONFilePath)
		if err != nil && !os.IsNotExist(err) {
			logrus.Errorf("Error reading existing JSON data from %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}

		// Append server JSON data to existing data
		allJSONData := append(existingJSONData, serverJSONData...)
		err = tools.AppendParsedDataToFile(serverJSONData, outputJSONFilePath)
		if err != nil {
			logrus.Errorf("Error appending server JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
			logrus.Errorf("Error writing server JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		err = tools.FileParserFromYAMLConfigServer(yamlFileparserFilePath, outputJSONFilePath)
		if err != nil {
			logrus.Errorf("Error parsing files at the server level: %v", err)
			os.Exit(1)
		}
		logrus.Infof("Server commands executed and output appended to %s.", outputJSONFilePath)

	}

	// Check if the instance argument is provided
	if instanceArg != "" {
		instanceCommands, err := tools.ReadInstanceCommandsFromYAML(yamlCommandsFilePath, instanceArg)
		if err != nil {
			logrus.Errorf("Error reading instance commands from YAML: %v", err)
			os.Exit(1)
		}

		base64InstanceOutputs, err := tools.ExecuteAndEncodeCommands(instanceCommands)
		if err != nil {
			logrus.Errorf("Error executing instance commands: %v", err)
			os.Exit(1)
		}

		// Create JSON data for instance commands
		var instanceJSONData []tools.JSONData
		for i, cmd := range instanceCommands {
			instanceJSONData = append(instanceJSONData, tools.JSONData{
				Command:     cmd.Command,
				Description: cmd.Description,
				Output:      base64InstanceOutputs[i],
				MonitorTag:  cmd.MonitorTag,
			})
		}

		// Get the existing JSON data from the file (if it exists)
		existingJSONData, err := tools.ReadJSONFromFile(outputJSONFilePath)
		if err != nil && !os.IsNotExist(err) {
			logrus.Errorf("Error reading existing JSON data from %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}

		// Append instance JSON data to existing data
		allJSONData := append(existingJSONData, instanceJSONData...)

		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
			logrus.Errorf("Error writing instance JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		// File parsing for the instance level
		err = tools.FileParserFromYAMLConfigInstance(yamlFileparserFilePath, outputJSONFilePath, instanceArg)
		if err != nil {
			logrus.Errorf("Error parsing files at the instance level: %v", err)
			os.Exit(1)
		}
		logrus.Infof("Instance commands executed and output appended to %s.", outputJSONFilePath)

	}

	if flag.NFlag() == 0 {
		flag.Usage()
	}

}
