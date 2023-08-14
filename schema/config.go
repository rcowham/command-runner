package schema

import (
	"fmt"
)

// Define default paths
const (
	DefaultCombineYAMLPath = "configs/combine.yaml"
	DefaultOutputJSONPath  = "/tmp/out.json"
)

var (
	OutputJSONFilePath  = DefaultOutputJSONPath
	YamlCombineFilePath = DefaultCombineYAMLPath
)

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

// CombineConfig represents the entire structure of combine.yaml
type CombineConfig struct {
	//	Files            []FileConfig `yaml:",inline"`
	Files            []FileConfig `yaml:"files"`
	InstanceCommands []Command    `yaml:"instance_commands"`
	ServerCommands   []Command    `yaml:"server_commands"`
}

// FileConfig represents each file configuration in combine.yaml
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

// CommandConfig holds the configuration from the YAML file for instance_commands and server_commands
// However, with the inclusion in CombineConfig, you might not need to use this separately.
type CommandConfig struct {
	InstanceCommands []Command `yaml:"instance_commands"`
	ServerCommands   []Command `yaml:"server_commands"`
}

/* Not feelings this here
type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}
*/
