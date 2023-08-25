// file_io.go
package tools

import (
	"command-runner/schema"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
)

// Function to write JSON data to a file with indentation for human-readability
func WriteJSONToFile(data []JSONData, OutputJSONFilePath string) error {
	logrus.Debugf("Writing JSON data to file: %s", OutputJSONFilePath)
	jsonString, err := json.MarshalIndent(data, "", "    ") // Use four spaces for indentation
	if err != nil {
		logrus.Errorf("Failed to marshal JSON data: %s", err)
		return err
	}

	// Check if the file exists
	if _, err := os.Stat(OutputJSONFilePath); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(OutputJSONFilePath)
		if err != nil {
			logrus.Errorf("Failed to create file %s: %s", OutputJSONFilePath, err)
			return err
		}
		defer file.Close()
	}

	return ioutil.WriteFile(OutputJSONFilePath, jsonString, 0644)
}

func ReadJSONFromFile(OutputJSONFilePath string) ([]JSONData, error) {
	// Check if the file exists before attempting to open it
	if _, err := os.Stat(OutputJSONFilePath); os.IsNotExist(err) {
		logrus.Warnf("File %s does not exist. Skipping read operation.", OutputJSONFilePath)
		return nil, err
	}

	logrus.Debugf("Reading JSON data from file: %s", OutputJSONFilePath)
	jsonFile, err := os.Open(OutputJSONFilePath)
	if err != nil {
		logrus.Warnf("Failed to open file %s: %s", OutputJSONFilePath, err) // changed from Errorf to Warnf
		return nil, err
	}
	defer jsonFile.Close()

	var jsonData []JSONData
	dec := json.NewDecoder(jsonFile)
	if err := dec.Decode(&jsonData); err != nil {
		logrus.Errorf("Failed to decode JSON data from file %s: %s", OutputJSONFilePath, err)
		return nil, err
	}

	return jsonData, nil
}

//func AppendParsedDataToFile(parsedData []JSONData, OutputJSONFilePath string) error {

func AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath stringintconv) error {
	logrus.Debugf("Appending parsed data to file: %s", OutputJSONFilePath)
	// Get the existing JSON data from the file (if it exists)
	existingJSONData, err := ReadJSONFromFile(OutputJSONFilePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Error reading existing JSON data from %s: %s", OutputJSONFilePath, err)
		return fmt.Errorf("error reading existing JSON data from %s: %s", OutputJSONFilePath, err)
	}

	// Append the new JSON data to the existing data
	allJSONData := append(existingJSONData, parsedData...)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(allJSONData, OutputJSONFilePath); err != nil {
		logrus.Errorf("Error appending parsed data to file %s: %s", OutputJSONFilePath, err)
		return fmt.Errorf("error writing JSON data to %s: %s", OutputJSONFilePath, err)
	}

	return nil
}
func appendExistingJSONData(newJSONData []JSONData, OutputJSONFilePath string) []JSONData {
	existingJSONData, err := ReadJSONFromFile(OutputJSONFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("File %s does not exist. Will create it.", OutputJSONFilePath)
			return newJSONData
		} else {
			logrus.Errorf("error reading existing JSON data: %s", err)
			os.Exit(1)
		}
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
