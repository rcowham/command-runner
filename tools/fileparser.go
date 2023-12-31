package tools

import (
	"command-runner/schema"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// TODO Description for instance file parsed FIX
// FileParserFromYAMLConfigOs reads a YAML configuration, parses the specified files at server level, and appends the
// results to the output JSON file.
// Returns an error if any issues arise during the parsing process.
func FileParserFromYAMLConfigOs(configFilePath string, OutputJSONFilePath string) error {
	config, err := readYAMLConfig(configFilePath)
	if err != nil {
		logrus.Errorf("error reading YAML config: %v", err)
		return fmt.Errorf("error reading YAML config: %w", err)
	}

	var hadError bool
	for _, file := range config.Files {
		filePath := file.PathToFile
		if file.ParsingLevel == "server" {
			if err := parseAndAppendAtOsLevel(filePath, file, OutputJSONFilePath); err != nil {
				logrus.Errorf("error parsing file %s: %v", filePath, err)
				hadError = true
				// don't return, continue with the next file
			}
		}
	}
	if hadError {
		return fmt.Errorf("encountered errors while parsing some files")
	}
	logrus.Info("Successfully parsed and appended data at OS level")
	return nil
}

// FileParserFromYAMLConfigP4 reads a YAML configuration, parses the specified files at instance level
// (replacing the instance placeholder in the file path with the provided instance name) and appends the results
// to the output file.
// Returns an error if any issues arise during the parsing process.
func FileParserFromYAMLConfigP4(configFilePath, OutputJSONFilePath, instance string) error {
	config, err := readYAMLConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading YAML config: %w", err)
	}

	var hadError bool
	for _, file := range config.Files {
		filePath := file.PathToFile
		if file.ParsingLevel == "instance" {
			filePath = strings.Replace(filePath, "%INSTANCE%", instance, 1)
			err := parseAndAppendAtP4Level(filePath, file, OutputJSONFilePath, instance)
			if err != nil {
				if os.IsNotExist(err) { // Check if error is because file does not exist
					logrus.Warnf("file %s does not exist: %v", filePath, err)
					hadError = true
				} else {
					logrus.Errorf("error parsing file %s: %v", filePath, err)
					hadError = true
				}
			}
		}
	}
	if hadError {
		logrus.Warn("Some files encountered errors during parsing. Please check the logs for more details.")
	}
	return nil
}

// parseAndAppendAtOsLevel is an internal function that takes in a filePath, its configuration and an output path.
// It parses the content based on the configuration and appends the result to the output file.
// TODO Refactor possibly
func parseAndAppendAtOsLevel(filePath string, fileConfig schema.FileConfig, OutputJSONFilePath string) error {
	parsedContent, err := parseContent(filePath, fileConfig)
	if err != nil {
		// If there's an error reading the file, handle it
		if os.IsNotExist(err) {
			logrus.Errorf("[OS] creating failed to parse for json")
			// File does not exist, append specific message to JSON
			message := fmt.Sprintf("File: %s was not found", filePath)
			jsonData := JSONData{
				Command:     "[OS] Failed to parse: " + filePath,
				Description: fmt.Sprintf("File: %v", filePath),
				Output:      EncodeToBase64(message),
				MonitorTag:  fileConfig.MonitorTag,
			}
			if err := AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath); err != nil {
				logrus.Errorf("[OS] error appending not found file data to output: %v", err)
			}
			// Now continue with the loop
			return nil
		}
		return err
	}
	return appendParsedData(filePath, parsedContent, fileConfig, OutputJSONFilePath)
}

// parseAndAppendAtP4Level is similar to parseAndAppendAtOsLevel, but it's specifically for parsing at the instance level.
// TODO Refactor possibly
func parseAndAppendAtP4Level(filePath string, fileConfig schema.FileConfig, OutputJSONFilePath, instanceArg string) error {
	parsedContent, err := parseContent(filePath, fileConfig)
	if err != nil {
		// If there's an error reading the file, handle it
		if os.IsNotExist(err) {
			logrus.Errorf("[P4] creating failed to parse for json")
			// File does not exist, append specific message to JSON
			message := fmt.Sprintf("File: %s was not found", filePath)
			jsonData := JSONData{
				Command:     "[P4] Failed to parse: " + filePath,
				Description: fmt.Sprintf("File: %v", filePath),
				Output:      EncodeToBase64(message),
				MonitorTag:  fileConfig.MonitorTag,
			}
			if err := AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath); err != nil {
				logrus.Errorf("[P4] error appending not found file data to output: %v", err)
			}
			// Now continue with the loop
			return nil
		}
		return err
	}
	return appendParsedData(filePath, parsedContent, fileConfig, OutputJSONFilePath)
}

// parseContent is an internal function that reads the content from a file based on the provided configuration.
// It looks for specific keywords to parse the content or returns the entire content if ParseAll is true.
func parseContent(filePath string, fileConfig schema.FileConfig) (string, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("failed to read file: %q: %v", filePath, err)
		//	return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
		return "", err // Return the original error

	}
	content := string(fileContent)

	// If ParseAll is true, return the full content (but still sanitize if needed)
	if fileConfig.ParseAll {
		logrus.Infof("Parsing entire content of file: %q", filePath)
		return sanitizeOutput(content, fileConfig.SanitizationKeywords), nil
	}

	var outputLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, keyword := range fileConfig.Keywords {
			if strings.Contains(line, keyword) {
				outputLines = append(outputLines, line)
				break
			}
		}
	}
	logrus.Infof("Parsed content from file: %q based on provided keywords", filePath)

	return sanitizeOutput(strings.Join(outputLines, "\n"), fileConfig.SanitizationKeywords), nil
}

// sanitizeOutput removes lines containing any of the provided sanitization keywords from the output.
func sanitizeOutput(output string, sanitizationKeywords []string) string {
	var sanitizedOutputLines []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		isSanitized := false
		for _, keyword := range sanitizationKeywords {
			if strings.Contains(line, keyword) {
				isSanitized = true
				break
			}
		}
		if !isSanitized {
			sanitizedOutputLines = append(sanitizedOutputLines, line)
		}
	}
	return strings.Join(sanitizedOutputLines, "\n")
}

// readYAMLConfig is an internal function to read and unmarshal the YAML configuration from a given file path.
func readYAMLConfig(configFilePath string) (*schema.FileParserConfig, error) {
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("Failed to read YAML config file: %v", err)
		return nil, fmt.Errorf("failed to read YAML config file: %w", err)
	}

	var config schema.FileParserConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		logrus.Errorf("failed to unmarshal YAML content: %v", err)
		return nil, fmt.Errorf("failed to unmarshal YAML content: %w", err)
	}

	return &config, nil
}

// appendParsedData takes the parsed content and appends it in a structured format to the provided output file.
func appendParsedData(filePath string, parsedContent string, fileConfig schema.FileConfig, OutputJSONFilePath string) error {
	// Now sanitize the parsed content
	sanitizedOutput := sanitizeOutput(parsedContent, fileConfig.SanitizationKeywords)

	jsonData := JSONData{
		Command:     "File parsed: " + filePath,
		Description: fmt.Sprintf("File: %v", filePath),
		Output:      EncodeToBase64(sanitizedOutput),
		MonitorTag:  fileConfig.MonitorTag,
	}

	return AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath)

}
