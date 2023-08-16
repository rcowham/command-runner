// gcp.go
//

package tools

// Add your GCP-specific functions and structures here.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetGCPInstanceIdentityInfo retrieves the instance identity document and tags from the AWS metadata service.
func GetGCPInstanceIdentityInfo(outputFilePath string) error {
	//Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true
	documentURL := "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true"
	documentOUT, err := getGCPEndpoint(documentURL)
	logrus.Info("Fetching GCP instance identity document...")

	if err != nil {
		logrus.Errorf("Failed to fetch GCP instance identity document: %s", err)
		return err
	}
	// Sanitize sensitive information from documentOUT
	sanitizedDocument, err := sanitizeGCPInstanceDocument(documentOUT)
	if err != nil {
		logrus.Errorf("Failed to sanitize GCP instance identity document: %s", err)
		return err
	}

	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(outputFilePath)
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Failed to read JSON from file %s: %s", outputFilePath, err)
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
	if err := WriteJSONToFile(existingJSONData, outputFilePath); err != nil {
		logrus.Errorf("Failed to write JSON to file %s: %s", outputFilePath, err)
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
