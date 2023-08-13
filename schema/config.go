package schema

//"gopkg.in/yaml.v2"

// Define default paths
const (
	DefaultCombineYAMLPath = "configs/combine.yaml"
	DefaultOutputJSONPath  = "/tmp/out.json"
)

var (
	OutputJSONFilePath  = DefaultOutputJSONPath
	YamlCombineFilePath = DefaultCombineYAMLPath
)

type FileParserConfig struct {
	Files []FileConfig `yaml:"files"`
}

// CombineConfig represents the entire structure of combine.yaml
type CombineConfig struct {
	Files            []FileConfig `yaml:",inline"`
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

/*
type JSONData struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Output      string `json:"output"`
	MonitorTag  string `json:"monitor_tag"`
}
*/
