package tools

import (
	"command-runner/schema"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func FileParserFromYAMLConfigServer(configFilePath string, outputJSONFilePath string) error {
	config, err := readYAMLConfig(configFilePath)
	if err != nil {
		logrus.Errorf("error reading YAML config: %v", err)
		return fmt.Errorf("error reading YAML config: %w", err)
	}

	for _, file := range config.Files {
		filePath := file.PathToFile
		if file.ParsingLevel == "server" {
			if err := parseAndAppendAtServerLevel(filePath, file, outputJSONFilePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func FileParserFromYAMLConfigInstance(configFilePath, outputFilePath, instance string) error {
	config, err := readYAMLConfig(configFilePath)
	if err != nil {

		return fmt.Errorf("error reading YAML config: %w", err)
	}

	for _, file := range config.Files {
		filePath := file.PathToFile
		if file.ParsingLevel == "instance" {
			filePath = strings.Replace(filePath, "%INSTANCE%", instance, 1)
			if err := parseAndAppendAtInstanceLevel(filePath, file, outputFilePath, instance); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseAndAppendAtServerLevel(filePath string, fileConfig schema.FileConfig, outputFilePath string) error {
	parsedContent, err := parseContent(filePath, fileConfig)
	if err != nil {
		return err
	}

	return appendParsedData(filePath, parsedContent, fileConfig, outputFilePath)
}

func parseAndAppendAtInstanceLevel(filePath string, fileConfig schema.FileConfig, outputFilePath, instanceArg string) error {
	parsedContent, err := parseContent(filePath, fileConfig)
	if err != nil {
		return err
	}

	return appendParsedData(filePath, parsedContent, fileConfig, outputFilePath)
}

func parseContent(filePath string, fileConfig schema.FileConfig) (string, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Errorf("failed to read file: %q: %v", filePath, err)
		return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
	}
	content := string(fileContent)

	// If ParseAll is true, return the full content (but still sanitize if needed)
	if fileConfig.ParseAll {
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

	return sanitizeOutput(strings.Join(outputLines, "\n"), fileConfig.SanitizationKeywords), nil
}
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

func appendParsedData(filePath string, parsedContent string, fileConfig schema.FileConfig, outputFilePath string) error {
	// Now sanitize the parsed content
	sanitizedOutput := sanitizeOutput(parsedContent, fileConfig.SanitizationKeywords)

	jsonData := JSONData{
		Command:     "File parsed: " + filePath,
		Description: fmt.Sprintf("File: %v", filePath),
		Output:      EncodeToBase64(sanitizedOutput),
		MonitorTag:  fileConfig.MonitorTag,
	}
	return AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath)
}
