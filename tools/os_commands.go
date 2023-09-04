package tools

import (
	"command-runner/schema"
	"fmt"

	"github.com/sirupsen/logrus"
)

// HandleOsCommands handles execution of OS level commands and file parsing
func HandleOsCommands(cloudProvider, OutputJSONFilePath string) error {
	err := HandleCloudProviders(cloudProvider, OutputJSONFilePath)
	if err != nil {
		// Log the error but continue
		logrus.Errorf("Error handling cloud provider %s: %v", cloudProvider, err)
	}

	// Rest of the code remains the same
	logrus.Info("Executing OS commands...")

	osCommands, err := ReadOsCommandsFromYAML(schema.DefaultCmdConfigYAMLPath) //TODO Fix this
	if err != nil {
		return fmt.Errorf("failed to read OS commands from YAML: %w", err)
	}

	base64OScmdsOutputs, err := ExecuteAndEncodeCommands(osCommands, false, "")
	if err != nil {
		return fmt.Errorf("failed to execute and encode commands: %w", err)
	}

	osJSONData := createJSONDataForCommands(osCommands, base64OScmdsOutputs)
	allJSONData := appendExistingJSONData(osJSONData, OutputJSONFilePath)

	if err := WriteJSONToFile(allJSONData, OutputJSONFilePath); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	if err := FileParserFromYAMLConfigOs(schema.DefaultCmdConfigYAMLPath, OutputJSONFilePath); err != nil { //TODO FIX THIS
		return fmt.Errorf("failed to parse file from YAML config: %w", err)
	}

	logrus.Infof("OS commands executed and output appended to %s.", OutputJSONFilePath)
	return nil
}
