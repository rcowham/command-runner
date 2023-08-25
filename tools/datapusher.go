package tools

import (
	"bytes"
	"command-runner/schema"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	//http constants
	maxIterations = 5
	//autoCloudTimeout = 5 * time.Second // assuming 5 seconds for the timeout
)

func PushToDataPushGateway(OutputJSONFilePath string, configFilePath string) error {
	config, err := schema.ParseMetricsConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error parsing metrics config: %w", err)
	}
	client := &http.Client{
		Timeout: autoCloudTimeout,
	}
	// Change the port from :9091 to :9092
	parsedURL := strings.Replace(config.Host, ":9091", ":9092", 1)
	config.Host = parsedURL

	iterations := 0
	STATUS := 1
	for STATUS != 0 {
		time.Sleep(1 * time.Second)
		iterations++

		logrus.Info("Pushing Support data")

		data, err := ioutil.ReadFile(OutputJSONFilePath)
		if err != nil {
			return fmt.Errorf("error reading tempLog: %w", err)
		}

		//		req, err := http.NewRequest("POST", fmt.Sprintf("%s/json/?customer=%s&instance=%s", metricsHost, metricsCustomer, metricsInstance), bytes.NewBuffer(data))
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/json/?customer=%s&instance=%s", config.Host, config.Customer, config.Instance), bytes.NewBuffer(data))

		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}
		//		req.SetBasicAuth(metricsUser, metricsPasswd)
		req.SetBasicAuth(config.User, config.Passwd)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		result := string(body)

		logrus.Infof("Checking result: %s", result)

		if result == `{"message":"invalid username or password"}` {
			STATUS = 1
			logrus.Warn("Retrying due to temporary password failure")
		} else {
			STATUS = 0
		}

		if iterations >= maxIterations {
			logrus.Error("Push loop iterations exceeded")
			return fmt.Errorf("push loop iterations exceeded")
		}
	}
	return nil
}
