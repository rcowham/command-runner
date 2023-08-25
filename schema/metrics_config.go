package schema

import (
	"bufio"
	"os"
	"strings"
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
