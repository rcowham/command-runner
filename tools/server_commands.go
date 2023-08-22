package tools

import (
	"command-runner/schema"
	"fmt"

	"github.com/sirupsen/logrus"
)

// HandleServerCommands handles execution of server commands and file parsing
func HandleServerCommands(cloudProvider, outputJSONFilePath string) error {
	err := HandleCloudProviders(cloudProvider, outputJSONFilePath)
	if err != nil {
		// Log the error but continue
		logrus.Errorf("Error handling cloud provider %s: %v", cloudProvider, err)
	}

	// Rest of the code remains the same
	logrus.Info("Executing server commands...")

	serverCommands, err := ReadServerCommandsFromYAML(schema.YamlCmdConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to read server commands from YAML: %w", err)
	}

	base64ServerOutputs, err := ExecuteAndEncodeCommands(serverCommands)
	if err != nil {
		return fmt.Errorf("failed to execute and encode commands: %w", err)
	}

	serverJSONData := createJSONDataForCommands(serverCommands, base64ServerOutputs)
	allJSONData := appendExistingJSONData(serverJSONData)

	if err := WriteJSONToFile(allJSONData, schema.OutputJSONFilePath); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	if err := FileParserFromYAMLConfigServer(schema.YamlCmdConfigFilePath, schema.OutputJSONFilePath); err != nil {
		return fmt.Errorf("failed to parse file from YAML config server: %w", err)
	}

	logrus.Infof("Server commands executed and output appended to %s.", schema.OutputJSONFilePath)
	return nil
}
