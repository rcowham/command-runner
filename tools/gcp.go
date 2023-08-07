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
)

// GetGCPInstanceIdentityInfo retrieves the instance identity document and tags from the AWS metadata service.
func GetGCPInstanceIdentityInfo(outputFilePath string) error {
	//Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true
	documentURL := "http://metadata.google.internal/computeMetadata/v1/instance/?recursive=true"
	documentOUT, err := getGCPEndpoint(documentURL)
	if err != nil {
		return err
	}
	// Sanitize sensitive information from documentOUT
	sanitizedDocument, err := sanitizeGCPInstanceDocument(documentOUT)
	if err != nil {
		return err
	}

	fmt.Println("Sanitized Instance Identity Document:")
	fmt.Println(string(sanitizedDocument)) // Debug print to see the sanitized document content

	fmt.Println("Instance Identity Document Raw:")
	fmt.Println(string(documentOUT)) // Debug print to see the raw documentOUT content

	// Get the existing JSON data from the file
	existingJSONData, err := ReadJSONFromFile(outputFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Append the Base64 encoded sanitizedDocument to the JSON data
	existingJSONData = append(existingJSONData, JSONData{
		Command:     "Instance Identity Document",
		Description: "GCP Instance Identity Document",
		Output:      EncodeToBase64(string(sanitizedDocument)),
		MonitorTag:  "GCP",
	})

	existingJSONData = append(existingJSONData)

	// Write the updated JSON data back to the file
	if err := WriteJSONToFile(existingJSONData, outputFilePath); err != nil {
		return err
	}

	return nil
}
func sanitizeGCPInstanceDocument(documentOUT []byte) ([]byte, error) {
	// Unmarshal JSON into a map
	var documentMap map[string]interface{}
	if err := json.Unmarshal(documentOUT, &documentMap); err != nil {
		return nil, err
	}

	// Remove the "ssh-keys" field from the map
	delete(documentMap["attributes"].(map[string]interface{}), "ssh-keys")

	// Marshal the modified map back into JSON
	sanitizedDocument, err := json.Marshal(documentMap)
	if err != nil {
		return nil, err
	}

	return sanitizedDocument, nil
}
func getGCPEndpoint(url string) ([]byte, error) {
	// Clean the URL to remove unwanted characters
	url = strings.TrimSpace(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		// If the response is 404, return the content as is without treating it as an error
		return body, nil
	} else if resp.StatusCode != http.StatusOK {
		// If the response status code is not 200 OK or 404 Not Found, return an error
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return body, nil
}
