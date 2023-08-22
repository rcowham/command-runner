// execution.go
package tools

import (
	"bytes"
	"command-runner/schema"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Function to execute a shell command and capture its output and error streams
func ExecuteShellCommand(command string, prependSource bool, instanceArg string) (string, string, error) {
	if prependSource {
		command = fmt.Sprintf("source /p4/common/config/p4_%s.vars; %s", instanceArg, command)
	}

	logrus.Debugf("Executing shell command: %s", command)
	cmd := exec.Command("bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to execute command %s: %s", command, err)
		return stdout.String(), stderr.String(), err
	}

	// ***Not currently used but should be in coming versions.*** If prependSource is true, fetch environment variables after the command has executed.
	if prependSource {
		envVars := make(map[string]string)
		for _, envVar := range []string{"P4CONFIG", "P4PORT", "P4USER", "P4CLIENT", "P4TICKETS", "P4TRUST"} {
			envVars[envVar] = os.Getenv(envVar)
		}
		logrus.Debugf("Environment variables: %v", envVars)
	}

	return stdout.String(), stderr.String(), nil
}

// Function to execute commands and encode output to Base64
func ExecuteAndEncodeCommands(commands []schema.Command, prependSource bool, instanceArg string) ([]string, error) {
	var base64Outputs []string

	for _, cmd := range commands {
		logrus.Debugf("Encoding command: %s", cmd.Command)
		output, _, err := ExecuteShellCommand(cmd.Command, prependSource, instanceArg)
		if err != nil {
			logrus.Errorf("Error executing and encoding command %s: %s", cmd.Command, err)
			return nil, err
		}
		base64Output := EncodeToBase64(output)
		base64Outputs = append(base64Outputs, base64Output)
	}

	return base64Outputs, nil
}
