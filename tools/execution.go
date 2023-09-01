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
		command = fmt.Sprintf("source %sp4_%s.vars; %s", schema.DefaultP4VarDir, instanceArg, command)
	}

	logrus.Debugf("Executing shell command: %s", command)
	cmd := exec.Command("bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to execute command: %s", command)
		logrus.Debugf("--Failed with error %s", err)
		logrus.Debugf("Returning %s, %s", stdout.String(), stderr.String())
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
	logrus.Debugf("Returning %s, %s", stdout.String(), stderr.String())
	return stdout.String(), stderr.String(), nil
}

// runCommand runs the given command and returns its output
// TODO combine with above ExecuteShellCommand later
func RunAutoBotCommand(cmdPath string, instanceArg string, prepend bool) (string, error) {
	prependSourceCmd := ""
	if prepend {
		prependSourceCmd = fmt.Sprintf("source %sp4_%s.vars; ", schema.DefaultP4VarDir, instanceArg) //TODO CLEAN UP
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

// Function to execute commands and encode output to Base64
func ExecuteAndEncodeCommands(commands []schema.Command, prependSource bool, instanceArg string) ([]string, error) {
	var base64Outputs []string

	for _, cmd := range commands {
		logrus.Debugf("Execute And Encode Command: %s", cmd.Command)
		output, stderrOutput, err := ExecuteShellCommand(cmd.Command, prependSource, instanceArg)
		if err != nil {
			logrus.Errorf("Error executing and encoding command %s: %s", cmd.Command, err)

			// Create a formatted error message string that includes the stderr output
			//errorMsg := fmt.Sprintf("[Instance: %s] Error processing %s: %s\n%s", instanceArg, cmd.Command, err, stderrOutput)
			var errorMsg string
			if instanceArg != "" {
				errorMsg = fmt.Sprintf("[Instance: %s] Error processing %s: %s\n%s", instanceArg, cmd.Command, err, stderrOutput)
			} else {
				errorMsg = fmt.Sprintf("Error processing %s: %s\n%s", cmd.Command, err, stderrOutput)
			}
			logrus.Errorf("errorMsg: %s", errorMsg)
			// Encode the error message string
			base64ErrorOutput := EncodeToBase64(errorMsg)

			// Append the encoded error message to the outputs
			base64Outputs = append(base64Outputs, base64ErrorOutput)

			continue // Move on to the next command
		}

		// For successful command execution, encode the output

		successMsg := output

		base64Output := EncodeToBase64(successMsg)
		base64Outputs = append(base64Outputs, base64Output)

	}
	logrus.Debug("ExecuteAndEncodeCommand completed")
	return base64Outputs, nil
}
