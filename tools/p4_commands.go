package tools

import (
	"command-runner/schema"
	"fmt"
)

// HandleInstanceCommands handles execution of instance commands and file parsing
func HandleP4Commands(instanceArg, outputJSONFilePath string) error {
	p4Commands, err := ReadP4CommandsFromYAML(schema.YamlCmdConfigFilePath, instanceArg)
	if err != nil {
		return fmt.Errorf("failed to read P4 commands from YAML: %w", err)
	}

	base64P4Outputs, err := ExecuteAndEncodeCommands(p4Commands, true, instanceArg)
	if err != nil {
		return fmt.Errorf("failed to execute and encode P4 commands: %w", err)
	}

	p4JSONData := createJSONDataForCommands(p4Commands, base64P4Outputs)
	allJSONData := appendExistingJSONData(p4JSONData)

	if err := WriteJSONToFile(allJSONData, outputJSONFilePath); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	if err := FileParserFromYAMLConfigP4(schema.YamlCmdConfigFilePath, outputJSONFilePath, instanceArg); err != nil {
		return fmt.Errorf("failed to parse file from YAML config (instance): %w", err)
	}

	return nil
}
