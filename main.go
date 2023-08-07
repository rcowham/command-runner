/*
TODO Error parsing files at the instance level breaks it so if it cant parse the file its stops.
*/
package main

import (
	"command-runner/tools"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

func init() {
	flag.StringVar(&outputJSONFilePath, "output", "out.json", "Path to the output JSON file")

	flag.StringVar(&cloudProvider, "cloud", "", "Cloud provider (aws, gcp, or azure)")

	flag.StringVar(&configFilePath, "config", "", "Path to the config file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	// Modify the outputJSONFilePath to write to the "output" directory
	//outputJSONFilePath = "output/out.json"

}

// Function to get the absolute path for a file or directory
func getAbsolutePath(baseDir, relPath string) (string, error) {
	absPath := filepath.Join(baseDir, relPath)
	return absPath, nil
}

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
			fmt.Println("Error getting executable path:", err)
			os.Exit(1)
		}

		configFilePath = filepath.Join(filepath.Dir(exePath), "configs", "config.yaml")
	}

	// Read the config file
	configFile, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding config file:", err)
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
	//, err := getAbsolutePath(configsDir, config.CommandsYAMLPath)
	if err != nil {
		fmt.Println("Error getting absolute path for commands.yaml:", err)
		os.Exit(1)
	}

	// If -cloud is provided, check if it's a valid cloud provider
	if cloudProvider != "" {
		switch cloudProvider {
		case "aws":
			//Logic to handle AWS-related functionality
			err := tools.GetAWSInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				fmt.Println("Error getting AWS instance identity info:", err)
				os.Exit(1)
			}
		case "gcp":
			//Logic to handle GCP-related functionality
			err := tools.GetGCPInstanceIdentityInfo(outputJSONFilePath)
			if err != nil {
				fmt.Println("Error getting GCP instance identity info:", err)
				os.Exit(1)
			}
		case "azure":
			// Add logic to handle Azure-related functionality
		default:
			fmt.Println("Error: Invalid cloud provider. Please specify aws, gcp, or azure.")
			os.Exit(1)
		}
	}

	// Check if the server argument is provided
	if serverArg {
		//		fmt.Println("Server ARG passed")
		// Execute and encode server commands
		serverCommands, err := tools.ReadServerCommandsFromYAML(yamlCommandsFilePath)
		if err != nil {
			fmt.Println("Error reading server commands from YAML:", err)
			os.Exit(1)
		}

		base64ServerOutputs, err := tools.ExecuteAndEncodeCommands(serverCommands)
		if err != nil {
			fmt.Println("Error executing server commands:", err)
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
			fmt.Printf("Error reading existing JSON data from %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}

		// Append server JSON data to existing data
		allJSONData := append(existingJSONData, serverJSONData...)
		err = tools.AppendParsedDataToFile(serverJSONData, outputJSONFilePath)
		if err != nil {
			fmt.Printf("Error appending server JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
			fmt.Printf("Error writing server JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		/*
			// Get the hostname of the server
			hostname, err := os.Hostname()
			if err != nil {
				fmt.Println("Error getting hostname:", err)
				os.Exit(1)
			}
		*/
		// File parsing for the server level
		//		fmt.Println("File Parser YAML Path (Server):", config.FileParserYAMLPath) // Print the file parser YAML path for server

		/* NEEDS TO BE LOOKED INTO SILLY APPENDING and file path issues
		err = tools.FileParserFromYAMLConfigServer(config.FileParserYAMLPath, outputJSONFilePath)
		if err != nil {
			fmt.Println("Error parsing files at the server level:", err)
			os.Exit(1)
		}
		fmt.Printf("%s Server commands executed and output appended to %s.\n", hostname, outputJSONFilePath)
		*/
	}

	// Check if the instance argument is provided
	if instanceArg != "" {
		//fmt.Println("Instance ARG passed")

		instanceCommands, err := tools.ReadInstanceCommandsFromYAML(yamlCommandsFilePath, instanceArg)
		if err != nil {
			fmt.Println("Error reading instance commands from YAML:", err)
			os.Exit(1)
		}

		base64InstanceOutputs, err := tools.ExecuteAndEncodeCommands(instanceCommands)
		if err != nil {
			fmt.Println("Error executing instance commands:", err)
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
			fmt.Printf("Error reading existing JSON data from %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}

		// Append instance JSON data to existing data
		allJSONData := append(existingJSONData, instanceJSONData...)

		// Write the updated JSON data back to the file
		if err := tools.WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
			fmt.Printf("Error writing instance JSON data to %s: %s\n", outputJSONFilePath, err)
			os.Exit(1)
		}
		// File parsing for the instance level
		//fmt.Println("File Parser YAML Path (Instance):", config.FileParserYAMLPath) // Print the file parser YAML path for instance
		err = tools.FileParserFromYAMLConfigInstance(yamlFileparserFilePath, outputJSONFilePath, instanceArg)
		if err != nil {
			fmt.Println("Error parsing files at the instance level:", err)
			os.Exit(1)
		}
		//fmt.Printf("Instance %s commands executed and output appended to %s.\n", instanceArg, outputJSONFilePath)
	}

	if flag.NFlag() == 0 {
		flag.Usage()
	}

	if flag.NFlag() == 0 {
		flag.Usage()
	}
}
