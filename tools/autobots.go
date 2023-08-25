package tools

import (
	"command-runner/schema"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	autobotsDir       string
	osScriptsExecuted bool
)

func init() {
	exeDir := schema.GetExecutableDir()
	autobotsDir = schema.GetConfigPath(exeDir, schema.AutobotsDir)
}

// HandleAutobotsScripts runs all scripts/binaries in the autobots directory
func HandleAutobotsScripts(outputFilePath string, instanceArg string, autobotsArg bool) error {
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
				continue
			}

			// For P4_ prefix, prepend is true; for OS_ prefix, prepend is false
			prepend := prefix == "P4_"
			output, err := runCommand(autobotsDir+"/"+file.Name(), instanceArg, prepend)

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
			logrus.Debugf("outputFilePath: %s", outputFilePath)

			// Save results to outputFilePath
			if err := AppendParsedDataToFile([]JSONData{jsonData}, outputFilePath); err != nil {
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

// runCommand runs the given command and returns its output
// TODO move this later
func runCommand(cmdPath string, instanceArg string, prepend bool) (string, error) {
	prependSourceCmd := ""
	if prepend {
		prependSourceCmd = fmt.Sprintf("source /p4/common/config/p4_%s.vars; ", instanceArg)
	}
	cmd := exec.Command("/bin/bash", "-c", prependSourceCmd+cmdPath)
	output, err := cmd.CombinedOutput()
	logrus.Debugf("Running script like so: %s", cmd)
	if err != nil {
		logrus.Errorf("Failed to execute %s: %s", cmdPath, err)
		return "", err
	}
	logrus.Debugf("Output of script %s", output)
	return string(output), nil
}

/*
// TODO move this later
func runCommandWithVars(cmdPath string, prependSource bool, instanceArg string) (string, error) {
	if prependSource {
		cmdPath = fmt.Sprintf("source /p4/common/config/p4_%s.vars; %s", instanceArg, cmdPath)
	}

	cmd := exec.Command("bash", "-c", cmdPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to execute %s: %s", cmdPath, err)
		return "", err
	}
	return string(output), nil
}
*/
