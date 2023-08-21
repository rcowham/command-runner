// execution.go
package tools

import (
	"bytes"
	"command-runner/schema"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Function to execute a shell command and capture its output and error streams
func ExecuteShellCommand(command string) (string, string, error) {
	logrus.Debugf("Executing shell command: %s", command)
	cmd := exec.Command("sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to execute command %s: %s", command, err)
		return stdout.String(), stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil
}

// Function to execute commands and encode output to Base64
func ExecuteAndEncodeCommands(commands []schema.Command) ([]string, error) {
	var base64Outputs []string

	for _, cmd := range commands {
		logrus.Debugf("Encoding command: %s", cmd.Command)
		output, _, err := ExecuteShellCommand(cmd.Command)
		if err != nil {
			logrus.Errorf("Error executing and encoding command %s: %s", cmd.Command, err)
			return nil, err
		}
		base64Output := EncodeToBase64(output)
		base64Outputs = append(base64Outputs, base64Output)
	}

	return base64Outputs, nil
}
