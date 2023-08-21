package tools

import (
	"command-runner/schema"
	"fmt"
)

// HandleInstanceCommands handles execution of instance commands and file parsing
func HandleInstanceCommands(instanceArg, outputJSONFilePath string) error {
	instanceCommands, err := ReadInstanceCommandsFromYAML(schema.YamlCombineFilePath, instanceArg)
	if err != nil {
		return fmt.Errorf("failed to read instance commands from YAML: %w", err)
	}

	base64InstanceOutputs, err := ExecuteAndEncodeCommands(instanceCommands)
	if err != nil {
		return fmt.Errorf("failed to execute and encode instance commands: %w", err)
	}

	instanceJSONData := createJSONDataForCommands(instanceCommands, base64InstanceOutputs)
	allJSONData := appendExistingJSONData(instanceJSONData)

	if err := WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	if err := FileParserFromYAMLConfigInstance(schema.YamlCombineFilePath, outputJSONFilePath, instanceArg); err != nil {
		return fmt.Errorf("failed to parse file from YAML config for instance: %w", err)
	}

	return nil
}
