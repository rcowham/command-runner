package schema

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type MetricsConfig struct {
	Host      string
	Customer  string
	Instance  string
	User      string
	Passwd    string
	CloudType string
}

func ParseMetricsConfig(filePath string) (MetricsConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return MetricsConfig{}, err
	}
	defer file.Close()

	config := MetricsConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "metrics_host":
			config.Host = value
		case "metrics_customer":
			config.Customer = value
		case "metrics_instance":
			config.Instance = value
		case "metrics_user":
			config.User = value
		case "metrics_passwd":
			config.Passwd = value
		case "metrics_cloudtype":
			config.CloudType = value
		}
	}

	if err := scanner.Err(); err != nil {
		return MetricsConfig{}, err
	}

	return config, nil
}
func UpdateMetricsConfig(metricsCloudType string) error {
	// Log the action
	logrus.Infof("Updating Metrics Config File with metrics_cloudtype=%s", metricsCloudType)

	// Read the current content of the configuration file
	content, err := os.ReadFile(MetricsConfigFile)
	if err != nil {
		logrus.Errorf("Error reading %s: %v", MetricsConfigFile, err)
		return fmt.Errorf("error reading %s: %v", MetricsConfigFile, err)
	}

	// Convert to string and check if the key already exists
	contentStr := string(content)
	if strings.Contains(contentStr, "metrics_cloudtype=") {
		// Replace the existing entry with the new value
		contentStr = strings.Replace(contentStr, "metrics_cloudtype="+metricsCloudType, "", -1)
		contentStr = strings.TrimSpace(contentStr) // Clean up possible extra newlines
		contentStr += fmt.Sprintf("\nmetrics_cloudtype=%s\n", metricsCloudType)
	} else {
		// Append the new entry if it doesn't exist
		contentStr += fmt.Sprintf("metrics_cloudtype=%s\n", metricsCloudType)
	}

	// Write the updated content back to the configuration file
	err = os.WriteFile(MetricsConfigFile, []byte(contentStr), 0644)
	if err != nil {
		logrus.Errorf("Error updating %s with metrics_cloudtype=%s: %v", MetricsConfigFile, metricsCloudType, err)
		return fmt.Errorf("error updating %s with metrics_cloudtype=%s: %v", MetricsConfigFile, metricsCloudType, err)
	}

	return nil
}
func IsCommandRunnerEnabled(MetricsConfigFile string) bool {
	file, err := os.Open(MetricsConfigFile)
	if err != nil {
		logrus.Errorf("Error opening metrics config file: %v", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "enabled=1" {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Errorf("Error reading metrics config file: %v", err)
	}

	return false
}
func GetCloudType(MetricsConfigFile string) (string, error) {
	config, err := ParseMetricsConfig(MetricsConfigFile)
	if err != nil {
		return "", err
	}

	if config.CloudType == "" {
		return "onprem", nil
	}

	return config.CloudType, nil
}

// FetchOrDetermineCloudProvider returns the cloud provider based on flags and configuration.
func FetchOrDetermineCloudProvider(autoCloudFlag bool, cloudProviderFlag string, metricsConfigFile string) string {

	if cloudProviderFlag == "onprem" { // Default value for the flag
		// Fetch from .push_metrics.cfg
		valueFromConfig, err := GetCloudType(metricsConfigFile)
		if err != nil {
			logrus.Warnf("Unable to get cloud type from metrics config, defaulting to onprem. Error: %v", err)
			return "onprem"
		}
		return valueFromConfig
	}

	return cloudProviderFlag
}
