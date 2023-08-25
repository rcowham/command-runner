package tools

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	//http constants
	autoCloudTimeout = 5 * time.Second // assuming 5 seconds for the timeout
)

func HandleCloudProviders(cloudProvider string, OutputJSONFilePath string) error {
	logrus.Infof("Cloud provider: %s", cloudProvider)
	switch cloudProvider {
	case "aws":
		return handleCloudProvider(GetAWSInstanceIdentityInfo, "AWS instance identity info", OutputJSONFilePath)
	case "gcp":
		return handleCloudProvider(GetGCPInstanceIdentityInfo, "GCP instance identity info", OutputJSONFilePath)
	case "azure":
		logrus.Warn("Azure cloud provider not yet implemented.")
		return nil // Nothing to do for Azure currently
	case "onprem":
		logrus.Warn("On-premises provider.")
		return nil // Nothing to do for on-prem currently
	default:
		logrus.Error("Invalid cloud provider. Please specify aws, gcp, azure, or onprem.")
		return fmt.Errorf("invalid cloud provider")
	}
}

func handleCloudProvider(providerFunc func(string) error, description string, OutputJSONFilePath string) error {
	err := providerFunc(OutputJSONFilePath)
	if err != nil {
		logrus.Errorf("error executing %s: %s", description, err)
		return err
	}
	return nil
}

// Save cloud handler errors
func saveErrorToJSON(OutputJSONFilePath, source, errorMessage, monitorTag string) error {
	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(OutputJSONFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create the concatenated description
	description := fmt.Sprintf("Error - %s: %s", monitorTag, source)

	// Append the error message to the JSON data
	errorJSON := JSONData{
		Command:     source,
		Description: description,
		Output:      EncodeToBase64(errorMessage),
		MonitorTag:  monitorTag,
	}

	existingJSONData = append(existingJSONData, errorJSON)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(existingJSONData, OutputJSONFilePath); err != nil {
		return err
	}

	return fmt.Errorf(errorMessage)
}
func DetectCloudProvider() (string, error) {
	logrus.Info("Detecting Cloud Provider")
	client := &http.Client{
		Timeout: autoCloudTimeout,
	}

	// Check Azure
	resp, err := client.Get("http://169.254.169.254/metadata/instance?api-version=2021-02-01")
	if err == nil && resp.StatusCode == http.StatusOK {
		logrus.Info("Azure Detected")
		return "azure", nil
	}

	// Check AWS
	resp, err = client.Get("http://169.254.169.254/latest/dynamic/instance-identity/document")
	if err == nil && resp.StatusCode == http.StatusOK {
		logrus.Info("AWS Detected")
		return "aws", nil
	}

	// Check GCP
	req, _ := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true", nil)
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err = client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		logrus.Info("GCP Detected")
		return "gcp", nil
	}
	logrus.Info("Unknown defaulting to onprem")
	return "onprem", nil
}
