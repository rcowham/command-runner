/*
 */
package main

import (
	"command-runner/schema"
	"command-runner/tools"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"

	"github.com/sirupsen/logrus"
)

var (
	cloudProvider      = kingpin.Flag("cloud", "Cloud provider (aws, gcp, or azure)").String()
	debug              = kingpin.Flag("debug", "Enable debug logging").Bool()
	OutputJSONFilePath = kingpin.Flag("output", "Path to the output JSON file").Default(schema.DefaultOutputJSONPath).String()
	instanceArg        = kingpin.Flag("instance", "Instance argument for the command-runner").String()
	serverArg          = kingpin.Flag("server", "Server argument for the command-runner").Bool()
	version            = "development"
)

func setupLogger(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()
	// Check if any flags were provided
	if !(*cloudProvider != "" || *debug || *OutputJSONFilePath != "" || *instanceArg != "" || *serverArg) {
		kingpin.Usage()
		return
	}
	// Setting up the logger
	setupLogger(*debug)

	// Determine the directory of the command-runner executable
	exePath, err := os.Executable()
	logrus.Debugf("Executable directory: %s", exePath)
	if err != nil {
		logrus.Error("Error getting executable path:", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	// Construct the path for combine.yaml
	schema.YamlCombineFilePath = filepath.Join(exeDir, schema.DefaultCombineYAMLPath)

	// Validate the combine.yaml file
	if err := schema.ValidateCombineYAML(schema.YamlCombineFilePath); err != nil {
		logrus.Fatal("Error validating combine.yaml:", err)
	}

	// Handle cloud providers
	if *cloudProvider != "" {
		logrus.Infof("Cloud provider: %s", *cloudProvider)
		switch *cloudProvider {
		case "aws":
			err := tools.GetAWSInstanceIdentityInfo(*OutputJSONFilePath)
			if err != nil {
				logrus.Error("Error getting AWS instance identity info:", err)
				os.Exit(1)
			}
		case "gcp":
			err := tools.GetGCPInstanceIdentityInfo(*OutputJSONFilePath)
			if err != nil {
				logrus.Error("Error getting GCP instance identity info:", err)
				os.Exit(1)
			}
		case "azure":
			// Add Azure handling logic here
			logrus.Warn("Azure cloud provider not yet implemented.")
			os.Exit(1)
		default:
			logrus.Error("Invalid cloud provider. Please specify aws, gcp, or azure.")
			os.Exit(1)
		}
	}

	// Handle server commands and file parsing
	if *serverArg {
		logrus.Info("Executing server commands...")
		// Replace ReadServerCommandsFromYAML with a new function that gets server commands from combine.yaml
		serverCommands, err := tools.ReadServerCommandsFromYAML(schema.YamlCombineFilePath)
		logrus.Debug("Parsed server commands: ", serverCommands)
		if err != nil {
			logrus.Fatalf("Error reading server commands from combine YAML: %v", err)
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
		existingJSONData, err := tools.ReadJSONFromFile(schema.OutputJSONFilePath)
		if err != nil && !os.IsNotExist(err) {
			logrus.Errorf("Error reading existing JSON data from %s: %s\n", schema.OutputJSONFilePath, err)
			os.Exit(1)
		}

		// Append server JSON data to existing data
		allJSONData := append(existingJSONData, serverJSONData...)
		err = tools.AppendParsedDataToFile(serverJSONData, schema.OutputJSONFilePath)
		if err != nil {
			logrus.Errorf("Error appending server JSON data to %s: %s\n", schema.OutputJSONFilePath, err)
			os.Exit(1)
		}
		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, schema.OutputJSONFilePath); err != nil {
			logrus.Errorf("Error writing server JSON data to %s: %s\n", schema.OutputJSONFilePath, err)
			os.Exit(1)
		}
		// For file parsing
		err = tools.FileParserFromYAMLConfigServer(schema.YamlCombineFilePath, schema.OutputJSONFilePath)
		if err != nil {
			logrus.Errorf("Error parsing files at the server level: %v", err)
			os.Exit(1)
		}
		logrus.Infof("Server commands executed and output appended to %s.", schema.OutputJSONFilePath)
	}

	// Handle instance commands and file parsing
	if *instanceArg != "" {
		logrus.Infof("Executing commands for instance: %s", *instanceArg)
		// Replace ReadInstanceCommandsFromYAML with a new function that gets instance commands from combine.yaml
		//OLD		instanceCommands, err := tools.ReadInstanceCommandsFromYAML(schema.YamlCombineFilePath, instanceArg)
		instanceCommands, err := tools.ReadInstanceCommandsFromYAML(schema.YamlCombineFilePath, *instanceArg)
		logrus.Debug("Parsed instance commands: ", instanceCommands)
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
		existingJSONData, err := tools.ReadJSONFromFile(schema.OutputJSONFilePath)
		if err != nil && !os.IsNotExist(err) {
			logrus.Errorf("Error reading existing JSON data from %s: %s\n", schema.OutputJSONFilePath, err)
			os.Exit(1)
		}

		// Append instance JSON data to existing data
		allJSONData := append(existingJSONData, instanceJSONData...)

		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, schema.OutputJSONFilePath); err != nil {
			logrus.Errorf("Error writing instance JSON data to %s: %s\n", schema.OutputJSONFilePath, err)
			os.Exit(1)
		}
		// For file parsing at the instance level
		err = tools.FileParserFromYAMLConfigInstance(schema.YamlCombineFilePath, schema.OutputJSONFilePath, *instanceArg)
		if err != nil {
			logrus.Errorf("Error parsing files at the instance level: %v", err)
			os.Exit(1)
		}
		logrus.Infof("Instance commands executed and output appended to %s.", schema.OutputJSONFilePath)
	}
	logrus.Info("Command-runner completed.")
}
