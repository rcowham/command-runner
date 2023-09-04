package tools

import (
	"command-runner/schema"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	autobotsDir       = schema.AutobotsDir //TODO Should be editable
	osScriptsExecuted bool
)

func init() {
	//	exeDir := schema.GetExecutableDir()
	//	autobotsDir = schema.GetConfigPath(exeDir, schema.AutobotsDir)
}

// HandleAutobotsScripts runs all scripts/binaries in the autobots directory
func HandleAutobotsScripts(OutputJSONFilePath string, instanceArg string, autobotsArg bool) error {
	files, err := ioutil.ReadDir(autobotsDir)
	if err != nil {
		return fmt.Errorf("error reading autobots directory: %w", err)
	}

	// First process OS_ prefixed files, then P4_ prefixed files
	prefixes := []string{"OS_", "P4_"}
	for _, prefix := range prefixes {
		// If we are about to process OS_ files and they've already been executed, skip
		if prefix == "OS_" && osScriptsExecuted {
			continue
		}

		for _, file := range files {
			if !isExecutable(file.Mode()) || !strings.HasPrefix(file.Name(), prefix) {
				// Skip files that don't match the current prefix
				logrus.Debugf("Skipping files that dont match prefix")
				continue
			}

			// For P4_ prefix, prepend is true; for OS_ prefix, prepend is false
			//prependSource := prefix == "P4_"
			prepend := prefix == "P4_"

			logrus.Debugf("running P4 Autobots")
			//cmdPath := filepath.Join(autobotsDir, file.Name())
			//output, _, err := ExecuteShellCommand(cmdPath, prependSource, instanceArg)
			output, err := RunAutoBotCommand(autobotsDir+"/"+file.Name(), instanceArg, prepend)
			//output, _, err := ExecuteShellCommand(cmd.Command, prependSource, instanceArg)
			if err != nil {
				logrus.Errorf("Error running command %s: %s", file.Name(), err)
				// Not returning here and instead proceeding to save the output
			}

			monitorTag := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			// Conditional Description based on the prefix
			var description string
			if prefix == "P4_" {
				description = fmt.Sprintf("[SDP Instance: %s] Output from %s", instanceArg, monitorTag)
			} else {
				description = fmt.Sprintf("[OS] Output from %s", monitorTag)
			}

			jsonData := JSONData{
				Command:     fmt.Sprintf("Autobot: %s", monitorTag),
				Description: description,
				Output:      EncodeToBase64(output),
				MonitorTag:  fmt.Sprintf("Autobot %s", monitorTag),
			}
			logrus.Debugf("results: %s", []JSONData{jsonData})
			logrus.Debugf("OutputJSONFilePath: %s", OutputJSONFilePath)

			// Save results to OutputJSONFilePath
			if err := AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath); err != nil {
				logrus.Errorf("[Autobots] error appending data to output for %s: %v", prefix, err)
			}
		}

		// Set the flag only after processing all OS-level scripts in one go
		if prefix == "OS_" {
			osScriptsExecuted = true
		}
	}

	logrus.Info("Autobots scripts executed and results saved.")
	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}

// Handle OS-level Autobots scripts
func HandleOSAutobotsScripts(OutputJSONFilePath string, instanceArg string) error {
	files, err := ioutil.ReadDir(autobotsDir)
	if err != nil {
		return fmt.Errorf("error reading autobots directory: %w", err)
	}

	for _, file := range files {
		if !isExecutable(file.Mode()) || !strings.HasPrefix(file.Name(), "OS_") {
			continue
		}

		cmdPath := filepath.Join(autobotsDir, file.Name())
		output, err := RunAutoBotCommand(cmdPath, instanceArg, false)
		if err != nil {
			logrus.Errorf("Error running OS-level command %s: %s", file.Name(), err)
		}

		description := fmt.Sprintf("[OS] Output from %s", strings.TrimPrefix(file.Name(), "OS_"))
		jsonData := JSONData{
			Command:     fmt.Sprintf("Autobot: %s", file.Name()),
			Description: description,
			Output:      EncodeToBase64(output),
			MonitorTag:  fmt.Sprintf("Autobot %s", file.Name()),
		}

		if err := AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath); err != nil {
			logrus.Errorf("[Autobots] error appending data to output for OS-level scripts: %v", err)
		}
	}

	logrus.Info("OS-level Autobots scripts executed and results saved.")
	return nil
}

// Handle SDP/P4-level Autobots scripts
func HandleSDPinstanceAutobotsScripts(OutputJSONFilePath string, instanceArg string) error {
	files, err := ioutil.ReadDir(autobotsDir)
	if err != nil {
		return fmt.Errorf("error reading autobots directory: %w", err)
	}

	for _, file := range files {
		if !isExecutable(file.Mode()) || !strings.HasPrefix(file.Name(), "P4_") {
			continue
		}

		cmdPath := filepath.Join(autobotsDir, file.Name())
		output, err := RunAutoBotCommand(cmdPath, instanceArg, true)
		if err != nil {
			logrus.Errorf("Error running SDP/P4-level command %s: %s", file.Name(), err)
		}

		description := fmt.Sprintf("[SDP Instance: %s] Output from %s", instanceArg, strings.TrimPrefix(file.Name(), "P4_"))
		jsonData := JSONData{
			Command:     fmt.Sprintf("Autobot: %s", file.Name()),
			Description: description,
			Output:      EncodeToBase64(output),
			MonitorTag:  fmt.Sprintf("Autobot %s", file.Name()),
		}

		if err := AppendParsedDataToFile([]JSONData{jsonData}, OutputJSONFilePath); err != nil {
			logrus.Errorf("[Autobots] error appending data to output for SDP/P4-level scripts: %v", err)
		}
	}

	logrus.Info("SDP/P4-level Autobots scripts executed and results saved.")
	return nil
}
