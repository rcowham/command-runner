package schema

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	ExeDir                   = GetExecutableDir()
	InstanceArg              = "1"
	DefaultCmdConfigYAMLPath string
	Vars2SourceFilePath      string
	CustomSourceVars         bool

// OLD	OutputJSONFilePath    = DefaultOutputJSONPath
// OLD YamlCmdConfigFilePath = DefaultCmdConfigYAMLPath
)

// Define default paths
const (
	P4baseDir         = "/p4" // TODO make this editable as well. and surely not /p4 consider flags
	LogFileName       = "command-runner.log"
	MainLogFilePath   = "/opt/perforce/command-runner/logs/" + LogFileName
	CmdConfigTamlPath = "configs/cmd_config.yaml"
	//ExeDefaultCmdConfigYAMLPath = ExeDir + CmdConfigTamlPath
	//DefaultCmdConfigYAMLPath = "configs/cmd_config.yaml"
	OutputJSONFilePath = "/tmp/out.json"
	AutobotsDir        = "autobots"
	MetricsConfigFile  = "/p4/common/config/.push_metrics.cfg"
	DefaultP4VarDir    = "/p4/common/config/"
	//		command = fmt.Sprintf("source %sp4_%s.vars; %s", schema.DefaultP4VarDir, instanceArg, command)
)

func init() {
	// Assuming ExeDir is a variable that you've defined and initialized somewhere.

	DefaultCmdConfigYAMLPath = ExeDir + CmdConfigTamlPath
	Vars2SourceFilePath = DefaultP4VarDir + "p4_" + InstanceArg + ".vars"
}
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		logrus.Fatal("Error getting executable path:", err)
	}
	return filepath.Dir(exePath)
}
func SetInstanceArg(arg string) {
	InstanceArg = arg
	//TODO Set Vars2Source?
}
func ReSetVars2SourceFilePath(arg string) {
	if !CustomSourceVars {
		// When CustomSourceVars is false
		logrus.Debugf("No custom vars file. Using: %s", Vars2SourceFilePath)
		InstanceArg = arg
		Vars2SourceFilePath = DefaultP4VarDir + "p4_" + arg + ".vars"
	} else {
		// When CustomSourceVars is true
		logrus.Debugf("Custom vars file. Using: %s", Vars2SourceFilePath)
	}
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
	Files      []FileConfig `yaml:"files"`
	P4Commands []Command    `yaml:"p4_commands"`
	OsCommands []Command    `yaml:"os_commands"`
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
	P4Commands []Command `yaml:"p4_commands"`
	OsCommands []Command `yaml:"os_commands"`
}
