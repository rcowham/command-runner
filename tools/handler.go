package tools

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func HandleCloudProviders(cloudProvider string, outputJSONFilePath string) error {
	logrus.Infof("Cloud provider: %s", cloudProvider)
	switch cloudProvider {
	case "aws":
		return handleCloudProvider(GetAWSInstanceIdentityInfo, "AWS instance identity info", outputJSONFilePath)
	case "gcp":
		return handleCloudProvider(GetGCPInstanceIdentityInfo, "GCP instance identity info", outputJSONFilePath)
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

func handleCloudProvider(providerFunc func(string) error, description string, outputJSONFilePath string) error {
	err := providerFunc(outputJSONFilePath)
	if err != nil {
		logrus.Errorf("error executing %s: %s", description, err)
		return err
	}
	return nil
}

// Save cloud handler errors
func saveErrorToJSON(outputFilePath, source, errorMessage, monitorTag string) error {
	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(outputFilePath)
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
	if err := WriteJSONToFile(existingJSONData, outputFilePath); err != nil {
		return err
	}

	return fmt.Errorf(errorMessage)
}
