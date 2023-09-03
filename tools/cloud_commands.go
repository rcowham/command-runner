/* cloudd_commands.go
 	aws_commands.go
	gcp_commands.go
*/

package tools

// Add your AWS-specific functions and structures here.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	AWSEndpoint   = "http://169.254.169.254"
	AWSTokenTTL   = "21600"
	ClientTimeout = 5 * time.Second
	//http constants
	autoCloudTimeout = 5 * time.Second // assuming 5 seconds for the timeout
)

var httpClient = &http.Client{Timeout: ClientTimeout}

// GetAWSToken retrieves the AWS metadata token.
func GetAWSToken(OutputJSONFilePath string) (string, error) {
	logrus.Info("Fetching AWS metadata token...")

	tokenURL := fmt.Sprintf("%s/latest/api/token", AWSEndpoint)
	req, err := http.NewRequest("PUT", tokenURL, nil)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSToken", fmt.Sprintf("Failed to create request for AWS token: %s", err), "AWS")
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", AWSTokenTTL)
	resp, err := httpClient.Do(req)

	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSToken", fmt.Sprintf("HTTP error while fetching token: %s", err), "AWS")
		return "", err
	}
	defer resp.Body.Close()

	token, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSToken", fmt.Sprintf("Failed to read response body: %s", err), "AWS")
		return "", err
	}

	// Check if the token length is zero
	if len(token) == 0 {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSToken", "Received empty AWS metadata token", "AWS")
		return "", fmt.Errorf("received empty AWS metadata token")
	}

	logrus.Info("Successfully fetched AWS metadata token.")
	return string(token), nil
}

// GetAWSInstanceIdentityInfo retrieves the instance identity document and tags from the AWS metadata service.
func GetAWSInstanceIdentityInfo(OutputJSONFilePath string) error {
	token, err := GetAWSToken(OutputJSONFilePath)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Failed to get AWS token: %s", err), "AWS")
		return err
	}

	documentURL := fmt.Sprintf("%s/latest/dynamic/instance-identity/document", AWSEndpoint)
	documentOUT, err := getAWSEndpoint(token, documentURL, OutputJSONFilePath)

	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Failed to get instance identity document: %s", err), "AWS")
		return err
	}
	logrus.Debug("Instance Identity Document Raw:")
	logrus.Debug(string(documentOUT))

	metadataURL := fmt.Sprintf("%s/latest/meta-data/tags/instance/", AWSEndpoint)
	metadataOUT, err := getAWSEndpoint(token, metadataURL, OutputJSONFilePath)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Failed to get metadata: %s", err), "AWS metadata")
		return err
	}
	logrus.Debug("Metadata Raw:")
	logrus.Debug(string(metadataOUT))

	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(OutputJSONFilePath)
	if err != nil && !os.IsNotExist(err) {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Error reading JSON from file: %s", err), "AWS")
		return err
	}

	// Append the Base64 encoded documentOUT and metadataOUT to the JSON data
	existingJSONData = append(existingJSONData, JSONData{
		Command:     "Instance Identity Document",
		Description: "AWS Instance Identity Document",
		Output:      EncodeToBase64(string(documentOUT)),
		MonitorTag:  "AWS",
	})

	metadataJSON := JSONData{
		Command:     "Metadata",
		Description: "AWS Metadata",
		Output:      EncodeToBase64(string(metadataOUT)),
		MonitorTag:  "AWS metadata",
	}

	existingJSONData = append(existingJSONData, metadataJSON)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(existingJSONData, OutputJSONFilePath); err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Error writing JSON to file: %s", err), "AWS")
		return err
	}

	return nil
}

func getAWSEndpoint(token, url, OutputJSONFilePath string) ([]byte, error) { // Added OutputJSONFilePath parameter
	url = strings.TrimSpace(url)

	logrus.Debugf("Fetching data from AWS endpoint: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "getAWSEndpoint", fmt.Sprintf("Failed to create request: %s", err), "AWS")
		return nil, err
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)
	resp, err := httpClient.Do(req)

	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "getAWSEndpoint", fmt.Sprintf("HTTP request failed for URL %s: %s", url, err), "AWS")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "getAWSEndpoint", fmt.Sprintf("Failed to read response body for URL %s: %s", url, err), "AWS")
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		// If the response is 404, return the content as is without treating it as an error
		return body, nil
	} else if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("Unexpected response status for URL %s: %s", url, resp.Status)
		saveErrorToJSON(OutputJSONFilePath, "getAWSEndpoint", errorMsg, "AWS")
		return nil, fmt.Errorf(errorMsg)
	}

	return body, nil
}

// gcp_commands.go
//

// Add your GCP-specific functions and structures here.

// GetGCPInstanceIdentityInfo retrieves the instance identity document and tags from the AWS metadata service.
func GetGCPInstanceIdentityInfo(OutputJSONFilePath string) error {
	documentURL := "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true"
	documentOUT, err := getGCPEndpoint(documentURL)
	logrus.Info("Fetching GCP instance identity document...")

	if err != nil {
		logrus.Errorf("Failed to fetch GCP instance identity document: %s", err)
		return saveErrorToJSON(OutputJSONFilePath, "Instance Identity Document", err.Error(), "GCP")
	}
	// Sanitize sensitive information from documentOUT
	sanitizedDocument, err := sanitizeGCPInstanceDocument(documentOUT)
	if err != nil {
		logrus.Errorf("Failed to sanitize GCP instance identity document: %s", err)
		return saveErrorToJSON(OutputJSONFilePath, "Instance Identity Document Sanitization", err.Error(), "GCP")
	}

	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(OutputJSONFilePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Failed to read JSON from file %s: %s", OutputJSONFilePath, err)
		return err
	}

	// Append the Base64 encoded sanitizedDocument to the JSON data
	existingJSONData = append(existingJSONData, JSONData{
		Command:     "Instance Identity Document",
		Description: "GCP Instance Identity Document",
		Output:      EncodeToBase64(string(sanitizedDocument)),
		MonitorTag:  "GCP",
	})
	logrus.Info("Appended GCP data to existing JSON.")

	// existingJSONData = append(existingJSONData)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(existingJSONData, OutputJSONFilePath); err != nil {
		logrus.Errorf("Failed to write JSON to file %s: %s", OutputJSONFilePath, err)
		return err
	}
	logrus.Info("Successfully updated JSON data with GCP instance information.")

	return nil
}
func sanitizeGCPInstanceDocument(documentOUT []byte) ([]byte, error) {
	logrus.Debug("Sanitizing GCP instance identity document...")

	// Unmarshal JSON into a map
	var documentMap map[string]interface{}
	if err := json.Unmarshal(documentOUT, &documentMap); err != nil {
		logrus.Errorf("Failed to unmarshal GCP document: %s", err)
		return nil, err
	}

	// Remove the "ssh-keys" field from the map
	delete(documentMap["attributes"].(map[string]interface{}), "ssh-keys")
	logrus.Debug("Removed ssh-keys from GCP document.")

	// Marshal the modified map back into JSON
	sanitizedDocument, err := json.Marshal(documentMap)
	if err != nil {
		logrus.Errorf("Failed to marshal sanitized GCP document: %s", err)
		return nil, err
	}

	return sanitizedDocument, nil
}
func getGCPEndpoint(url string) ([]byte, error) {
	logrus.Debugf("Fetching data from GCP endpoint: %s", url)

	// Clean the URL to remove unwanted characters
	url = strings.TrimSpace(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.Errorf("Failed to create request for GCP endpoint %s: %s", url, err)
		return nil, err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Failed to fetch data from GCP endpoint %s: %s", url, err)
		return nil, err
	}
	logrus.Debugf("Received response from GCP endpoint %s with status: %s", url, resp.Status)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read response body from GCP endpoint %s: %s", url, err)
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		// If the response is 404, return the content as is without treating it as an error
		return body, nil
	} else if resp.StatusCode != http.StatusOK {
		// If the response status code is not 200 OK or 404 Not Found, return an error
		logrus.Errorf("Unexpected response status from GCP endpoint %s: %s", url, resp.Status)
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return body, nil
}

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
