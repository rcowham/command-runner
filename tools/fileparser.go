package tools

import (
	"fmt"
	"io/ioutil"

	//	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type FileParserConfig struct {
	Files []struct {
		PathToFile           string   `yaml:"pathtofile"`
		Keywords             []string `yaml:"keywords"`
		ParseAll             bool     `yaml:"parseAll"`
		ParsingLevel         string   `yaml:"parsingLevel"`
		SanitizationKeywords []string `yaml:"sanitizationKeywords"`
	} `yaml:"files"`
}

type FileConfig struct {
	PathToFile           string   `yaml:"pathtofile"`
	Keywords             []string `yaml:"keywords"`
	ParseAll             bool     `yaml:"parseAll"`
	ParsingLevel         string   `yaml:"parsingLevel"`
	SanitizationKeywords []string `yaml:"sanitizationKeywords"`
}

// func FileParserFromYAMLConfigServer(configFilePath string) error {
func FileParserFromYAMLConfigServer(configFilePath string, outputJSONFilePath string) error {
	//	fmt.Println("I'm the fileParser config server")
	// Read and parse the YAML configuration file
	config, err := readYAMLConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading YAML config: %w", err)
	}

	// Parse each file according to the configuration
	for _, file := range config.Files {
		// Access the "PathToFile" field in the struct
		filePath := file.PathToFile

		// Check if parsingLevel is "server"
		if file.ParsingLevel == "server" {
			// Parse at the server level
			if err := parseAtServerLevel(filePath, file.Keywords, file.SanitizationKeywords, file.ParseAll, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

// Separate function for instance-level parsing
func FileParserFromYAMLConfigInstance(configFilePath, outputFilePath, instance string) error {
	//	fmt.Println("I'm the fileParser config instance")
	// Read and parse the YAML configuration file
	config, err := readYAMLConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading YAML config: %w", err)
	}

	// Parse each file according to the configuration
	for _, file := range config.Files {
		// Access the "PathToFile" field in the struct
		filePath := file.PathToFile

		// Check if parsingLevel is "instance"
		if file.ParsingLevel == "instance" {
			// Replace the "%INSTANCE%" placeholder with the actual instance name
			filePath = strings.Replace(filePath, "%INSTANCE%", instance, 1)
			// Parse at the instance level
			if err := parseAtInstanceLevel(filePath, file.Keywords, file.SanitizationKeywords, file.ParseAll, outputFilePath, instance); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseAtServerLevel(filePath string, keywords, sanitizationKeywords []string, parseAll bool, outputFilePath string) error {
	// Read the content of the file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	// Convert the content to a string
	content := string(fileContent)

	// Process all lines if parseAll is true and keywords is empty
	if len(keywords) == 0 && parseAll {
		//		fmt.Println("Server Level Output for", filePath)
		//fmt.Println(content)
		return nil
	}

	// Process the lines containing the keywords and print them
	// fmt.Println("Server Level Output for", filePath)
	var outputLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				// Print the line (you can perform additional operations here if needed)
				//fmt.Println(line)
				outputLines = append(outputLines, line)
				break // Move to the next line after processing the current keyword
			}
		}
	}
	// Sanitize the output before printing it
	//sanitizedOutput := sanitizeOutput(strings.Join(outputLines, "\n"), sanitizationKeywords)

	// Print the sanitized output
	//fmt.Println(sanitizedOutput)
	//fmt.Println("Based:", EncodeToBase64(sanitizedOutput))
	// Call the function to append parsed data for server level
	if err := AppendServerParsedData(filePath, keywords, sanitizationKeywords, parseAll, outputFilePath); err != nil {
		return err
	}
	return nil
}

func parseAtInstanceLevel(filePath string, keywords, sanitizationKeywords []string, parseAll bool, outputFilePath, instanceArg string) error {
	// Read the content of the file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	// Convert the content to a string
	content := string(fileContent)

	// Process all lines if parseAll is true and keywords is empty
	if len(keywords) == 0 && parseAll {
		//fmt.Println("Instance Level Output for", filePath)
		//fmt.Println(content)
		return nil
	}

	// Process the lines containing the keywords and print them
	//fmt.Println("Instance Level Output for", filePath)
	var outputLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				// Print the line (you can perform additional operations here if needed)
				//fmt.Println(line)
				outputLines = append(outputLines, line)
				break // Move to the next line after processing the current keyword
			}
		}
	}

	// Sanitize the output before printing it
	//sanitizedOutput := sanitizeOutput(strings.Join(outputLines, "\n"), sanitizationKeywords)

	// Print the sanitized output
	//fmt.Println("Sanitized Output:")
	//fmt.Println(sanitizedOutput)
	//fmt.Println("Based:", EncodeToBase64(sanitizedOutput))
	// Call the function to append parsed data for instance level
	if err := AppendInstanceParsedData(filePath, keywords, sanitizationKeywords, parseAll, outputFilePath, instanceArg); err != nil {
		return err
	}
	return nil
}

// Sanitize the output as needed before appending to the output file.
func sanitizeOutput(output string, sanitizationKeywords []string) string {
	// Add your sanitization logic here.
	// For example, you can remove lines containing the sanitization keywords from the output.
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

func readYAMLConfig(configFilePath string) (*FileParserConfig, error) {
	// Print the absolute path of the config file for debugging
	/*	absConfigPath, err := filepath.Abs(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path of config file: %w", err)
		}
		fmt.Println("Absolute path of the config file:", absConfigPath)
	*/
	// Read the content of the YAML configuration file
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML config file: %w", err)
	}

	// Parse the YAML content into the FileParserConfig struct
	var config FileParserConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
	}

	// Ensure that all necessary fields are provided in the YAML config
	if len(config.Files) == 0 {
		return nil, fmt.Errorf("files must be provided in the YAML config")
	}

	return &config, nil
}

// Function to append parsed data for server level to the JSON file
// func AppendServerParsedData(filePath string, keywords, sanitizationKeywords []string, parseAll bool, outputFilePath string) error {
func AppendServerParsedData(filePath string, keywords, sanitizationKeywords []string, parseAll bool, outputFilePath string) error {
	// Read the content of the file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	// Convert the content to a string
	content := string(fileContent)

	// Process all lines if parseAll is true and keywords is empty
	if len(keywords) == 0 && parseAll {
		// Create JSON data for the parsed file without parsing individual lines
		jsonData := JSONData{
			Command:     "File parsed: " + filePath,
			Description: fmt.Sprintf("Keywords: %v, ParseAll: %t, SanitizationKeywords: %v", keywords, parseAll, sanitizationKeywords),
			Output:      EncodeToBase64(content),
			MonitorTag:  "pathtofile: " + filePath,
		}

		// Append the JSON data to the JSON file
		err := AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath)
		if err != nil {
			return fmt.Errorf("error appending instance JSON data to %s: %w", outputFilePath, err)
		}

		return nil
	}

	// Process the lines containing the keywords and append the JSON data
	var outputLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				outputLines = append(outputLines, line)
				break // Move to the next line after processing the current keyword
			}
		}
	}

	// Sanitize the output before appending it to the file
	sanitizedOutput := sanitizeOutput(strings.Join(outputLines, "\n"), sanitizationKeywords)

	// Create JSON data for the parsed file
	jsonData := JSONData{
		Command:     "File parsed: " + filePath,
		Description: fmt.Sprintf("Keywords: %v, ParseAll: %t, SanitizationKeywords: %v", keywords, parseAll, sanitizationKeywords),
		Output:      EncodeToBase64(sanitizedOutput),
		MonitorTag:  "pathtofile: " + filePath,
	}

	// Append the JSON data to the JSON file
	err = AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath)
	if err != nil {
		return fmt.Errorf("error appending server JSON data to %s: %w", outputFilePath, err)
	}

	return nil
}

// Function to append parsed data for instance level to the JSON file
func AppendInstanceParsedData(filePath string, keywords, sanitizationKeywords []string, parseAll bool, outputFilePath, instanceArg string) error {
	// Read the content of the file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	// Convert the content to a string
	content := string(fileContent)

	// Process all lines if parseAll is true and keywords is empty
	if len(keywords) == 0 && parseAll {
		// Create JSON data for the parsed file without parsing individual lines
		jsonData := JSONData{
			Command:     "File parsed: " + filePath,
			Description: fmt.Sprintf("Keywords: %v, ParseAll: %t, SanitizationKeywords: %v", keywords, parseAll, sanitizationKeywords),
			Output:      EncodeToBase64(content),
			MonitorTag:  "pathtofile: " + filePath,
		}

		// Append the JSON data to the JSON file
		err := AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath)
		if err != nil {
			return fmt.Errorf("error appending instance JSON data to %s: %w", outputFilePath, err)
		}

		return nil
	}

	// Process the lines containing the keywords and append the JSON data
	var outputLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				outputLines = append(outputLines, line)
				break // Move to the next line after processing the current keyword
			}
		}
	}

	// Sanitize the output before appending it to the file
	sanitizedOutput := sanitizeOutput(strings.Join(outputLines, "\n"), sanitizationKeywords)

	// Create JSON data for the parsed file
	jsonData := JSONData{
		Command:     "File parsed: " + filePath,
		Description: fmt.Sprintf("Keywords: %v, ParseAll: %t, SanitizationKeywords: %v", keywords, parseAll, sanitizationKeywords),
		Output:      EncodeToBase64(sanitizedOutput),
		MonitorTag:  "pathtofile: " + filePath,
	}

	// Append the JSON data to the JSON file
	err = AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath)
	if err != nil {
		return fmt.Errorf("error appending instance JSON data to %s: %w", outputFilePath, err)
	}

	return nil
}
