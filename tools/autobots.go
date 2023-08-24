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

var autobotsDir string

func init() {
	exeDir := schema.GetExecutableDir()
	autobotsDir = schema.GetConfigPath(exeDir, schema.AutobotsDir)
}

// HandleAutobotsScripts runs all scripts/binaries in the autobots directory
func HandleAutobotsScripts(outputFilePath string) error {
	files, err := ioutil.ReadDir(autobotsDir)
	if err != nil {
		return fmt.Errorf("error reading autobots directory: %w", err)
	}

	var results []JSONData
	for _, file := range files {
		if !isExecutable(file.Mode()) {
			logrus.Infof("Skipping non-executable file: %s", file.Name())
			continue
		}

		output, err := runCommand(autobotsDir + "/" + file.Name())
		if err != nil {
			logrus.Errorf("Error running command %s: %s", file.Name(), err)
			// Not returning here and instead proceeding to save the output
		}

		monitorTag := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		results = append(results, JSONData{
			Command:     fmt.Sprintf("Autobot: %s", monitorTag),
			Description: fmt.Sprintf("Output from %s", monitorTag),
			Output:      EncodeToBase64(output), // Capturing the output regardless
			MonitorTag:  fmt.Sprintf("Autobot %s", monitorTag),
		})
	}

	// Save results to outputFilePath
	if err := AppendParsedDataToFile(results, outputFilePath); err != nil {
		return fmt.Errorf("error appending autobots data to file: %w", err)
	}

	logrus.Info("Autobots scripts executed and results saved.")

	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}

// runCommand runs the given command and returns its output
func runCommand(cmdPath string) (string, error) {
	cmd := exec.Command(cmdPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to execute %s: %s", cmdPath, err)
		return "", err
	}
	return string(output), nil
}
