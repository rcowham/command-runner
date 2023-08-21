// file_io.go
package tools

import (
	"command-runner/schema"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

// Function to write JSON data to a file with indentation for human-readability
func WriteJSONToFile(data []JSONData, filePath string) error {
	logrus.Debugf("Writing JSON data to file: %s", filePath)
	jsonString, err := json.MarshalIndent(data, "", "    ") // Use four spaces for indentation
	if err != nil {
		logrus.Errorf("Failed to marshal JSON data: %s", err)
		return err
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(filePath)
		if err != nil {
			logrus.Errorf("Failed to create file %s: %s", filePath, err)
			return err
		}
		defer file.Close()
	}

	return ioutil.WriteFile(filePath, jsonString, 0644)
}

func ReadJSONFromFile(filePath string) ([]JSONData, error) {
	logrus.Debugf("Reading JSON data from file: %s", filePath)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("Failed to open file %s: %s", filePath, err)
		return nil, err
	}
	defer jsonFile.Close()

	var jsonData []JSONData
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&jsonData); err != nil {
		logrus.Errorf("Failed to decode JSON data from file %s: %s", filePath, err)
		return nil, err
	}

	return jsonData, nil
}

func AppendParsedDataToFile(parsedData []JSONData, filePath string) error {
	logrus.Debugf("Appending parsed data to file: %s", filePath)
	// Get the existing JSON data from the file (if it exists)
	existingJSONData, err := ReadJSONFromFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Error reading existing JSON data from %s: %s", filePath, err)
		return fmt.Errorf("error reading existing JSON data from %s: %s", filePath, err)
	}

	// Append the new JSON data to the existing data
	allJSONData := append(existingJSONData, parsedData...)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(allJSONData, filePath); err != nil {
		logrus.Errorf("Error appending parsed data to file %s: %s", filePath, err)
		return fmt.Errorf("error writing JSON data to %s: %s", filePath, err)
	}

	return nil
}
func appendExistingJSONData(newJSONData []JSONData) []JSONData {
	existingJSONData, err := ReadJSONFromFile(schema.OutputJSONFilePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("error append existing JSON data: %s", err)
		os.Exit(1)
	}
	return append(existingJSONData, newJSONData...)
}
func createJSONDataForCommands(commands []schema.Command, outputs []string) []JSONData {
	var jsonData []JSONData
	for i, cmd := range commands {
		jsonData = append(jsonData, JSONData{
			Command:     cmd.Command,
			Description: cmd.Description,
			Output:      outputs[i],
			MonitorTag:  cmd.MonitorTag,
		})
	}
	return jsonData
}
