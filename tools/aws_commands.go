// aws_commands.go
//

package tools

// Add your AWS-specific functions and structures here.

import (
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
	//documentOUT, err := getAWSEndpoint(token, documentURL)
	documentOUT, err := getAWSEndpoint(token, documentURL, OutputJSONFilePath)

	if err != nil {
		saveErrorToJSON(OutputJSONFilePath, "GetAWSInstanceIdentityInfo", fmt.Sprintf("Failed to get instance identity document: %s", err), "AWS")
		return err
	}
	logrus.Debug("Instance Identity Document Raw:")
	logrus.Debug(string(documentOUT))

	metadataURL := fmt.Sprintf("%s/latest/meta-data/tags/instance/", AWSEndpoint)
	//metadataOUT, err := getAWSEndpoint(token, metadataURL)
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
