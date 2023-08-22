package schema

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Define default paths
const (
	DefaultCmdConfigYAMLPath = "configs/cmd_config.yaml"
	DefaultOutputJSONPath    = "/tmp/out.json"
)

var (
	OutputJSONFilePath    = DefaultOutputJSONPath
	YamlCmdConfigFilePath = DefaultCmdConfigYAMLPath
)

func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		logrus.Fatal("Error getting executable path:", err)
	}
	return filepath.Dir(exePath)
}

func GetConfigPath(basePath, configName string) string {
	return filepath.Join(basePath, configName)
}
func (fc *FileConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	mapped, ok := raw.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("expected a map")
	}

	for key, value := range mapped {
		strKey := fmt.Sprintf("%v", key)

		switch strKey {
		case "pathtofile":
			fc.PathToFile, _ = value.(string)
		case "monitor_tag":
			fc.MonitorTag, _ = value.(string)
		case "parseAll":
			fc.ParseAll, _ = value.(bool)
		case "parsingLevel":
			fc.ParsingLevel, _ = value.(string)
		case "sanitizationKeywords":
			if sk, ok := value.([]interface{}); ok {
				for _, s := range sk {
					fc.SanitizationKeywords = append(fc.SanitizationKeywords, fmt.Sprintf("%v", s))
				}
			}
		case "keywords":
			// This is where we handle both scenarios
			if kw, ok := value.([]interface{}); ok {
				// it's a list of keywords
				for _, k := range kw {
					fc.Keywords = append(fc.Keywords, fmt.Sprintf("%v", k))
				}
			} else if kw, ok := value.(string); ok {
				// it's a single keyword
				fc.Keywords = append(fc.Keywords, kw)
			}
		}
	}
	return nil
}

type FileParserConfig struct {
	Files []FileConfig `yaml:"files"`
}

// CmdConfig represents the entire structure of cmd_config.yaml
type CmdConfig struct {
	//	Files            []FileConfig `yaml:",inline"`
	Files []FileConfig `yaml:"files"`
	//	InstanceCommands []Command    `yaml:"instance_commands"`
	P4Commands []Command `yaml:"p4_commands"`
	//	ServerCommands []Command `yaml:"server_commands"`
	OsCommands []Command `yaml:"os_commands"`
}

// FileConfig represents each file configuration in cmd_config.yaml
type FileConfig struct {
	PathToFile           string   `yaml:"pathtofile"`
	Keywords             []string `yaml:"keywords"`
	ParseAll             bool     `yaml:"parseAll"`
	ParsingLevel         string   `yaml:"parsingLevel"`
	SanitizationKeywords []string `yaml:"sanitizationKeywords"`
	MonitorTag           string   `yaml:"monitor_tag"`
}

// Command represents individual command details
type Command struct {
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
	MonitorTag  string `yaml:"monitor_tag"`
}

// CommandConfig holds the configuration from the YAML file for p4_commands (formerly instance_commands) and os_commands(formerly server_commands)
// However, with the inclusion in CmdConfigConfig, you might not need to use this separately.
type CommandConfig struct {
	//	InstanceCommands []Command `yaml:"instance_commands"`
	P4Commands []Command `yaml:"p4_commands"`
	//	ServerCommands []Command `yaml:"server_commands"`
	OsCommands []Command `yaml:"os_commands"`
}

/* Not feelings this here
type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}
*/
