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
	//"gopkg.in/yaml.v2"
)

// Define default paths
const (
	defaultCombineYAMLPath = "configs/combine.yaml"
	defaultOutputJSONPath  = "/tmp/out.json"
)

var (
	outputJSONFilePath  = defaultOutputJSONPath
	yamlCombineFilePath = defaultCombineYAMLPath
	cloudProvider       string
	debug               bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&outputJSONFilePath, "output", defaultOutputJSONPath, "Path to the output JSON file")
	flag.StringVar(&cloudProvider, "cloud", "", "Cloud provider (aws, gcp, or azure)")

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

func main() {
	// Determine the directory of the command-runner executable
	exePath, err := os.Executable()
	if err != nil {
		logrus.Errorf("Error getting executable path: %v", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	// Construct the path for combine.yaml
	yamlCombineFilePath = filepath.Join(exeDir, defaultCombineYAMLPath)

	var instanceArg string
	var serverArg bool

	flag.StringVar(&instanceArg, "instance", "", "Instance argument for the command-runner")
	flag.BoolVar(&serverArg, "server", false, "Server argument for the command-runner")

	flag.Parse()

	// Handle cloud providers
	if cloudProvider != "" {
		switch cloudProvider {
		case "aws":
			err := tools.GetAWSInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				logrus.Errorf("Error getting AWS instance identity info: %v", err)
				os.Exit(1)
			}
		case "gcp":
			err := tools.GetGCPInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				logrus.Errorf("Error getting GCP instance identity info: %v", err)
				os.Exit(1)
			}
		case "azure":
			// Add Azure handling logic here
		default:
			logrus.Errorf("Error: Invalid cloud provider. Please specify aws, gcp, or azure.")
			os.Exit(1)
		}
	}

	// Handle server commands and file parsing
	if serverArg {
		// Replace ReadServerCommandsFromYAML with a new function that gets server commands from combine.yaml
		serverCommands, err := tools.ReadServerCommandsFromYAML(yamlCombineFilePath)
		if err != nil {
			logrus.Errorf("Error reading server commands from combine YAML: %v", err)
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
		// For file parsing
		err = tools.FileParserFromYAMLConfigServer(yamlCombineFilePath, outputJSONFilePath)
		if err != nil {
			logrus.Errorf("Error parsing files at the server level: %v", err)
			os.Exit(1)
		}
		logrus.Infof("Server commands executed and output appended to %s.", outputJSONFilePath)
	}

	// Handle instance commands and file parsing
	if instanceArg != "" {
		// Replace ReadInstanceCommandsFromYAML with a new function that gets instance commands from combine.yaml
		instanceCommands, err := tools.ReadInstanceCommandsFromYAML(yamlCombineFilePath, instanceArg)
		if err != nil {
			logrus.Errorf("Error reading instance commands from combine YAML: %v", err)
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
		// For file parsing at the instance level
		err = tools.FileParserFromYAMLConfigInstance(yamlCombineFilePath, outputJSONFilePath, instanceArg)
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
