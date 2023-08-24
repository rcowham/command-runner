/*
 */
package main

import (
	"command-runner/helpers"
	"command-runner/schema"
	"command-runner/tools"

	"os"

	"github.com/alecthomas/kingpin/v2"

	"github.com/sirupsen/logrus"
)

var (
	cloudProvider      = kingpin.Flag("cloud", "Cloud provider (aws, gcp, azure, or onprem)").Short('c').Default("onprem").String()
	debug              = kingpin.Flag("debug", "Enable debug logging").Short('d').Bool()
	OutputJSONFilePath = kingpin.Flag("output", "Path to the output JSON file").Short('o').Default(schema.DefaultOutputJSONPath).String()
	instanceArg        = kingpin.Flag("instance", "SDP instance commands argument for the command-runner").Short('i').String()
	serverArg          = kingpin.Flag("server", "OS commands argument for the command-runner").Short('s').Bool()
	autobotsArg        = kingpin.Flag("autobots", "Enable running autobots scripts").Short('a').Bool()
	version            = "development"
)

func validateFlags() bool {
	return isValidProvider() && isValidFlag()
}

func isValidProvider() bool {
	switch *cloudProvider {
	case "aws", "gcp", "azure", "onprem":
		return true
	default:
		logrus.Errorf("Invalid cloud provider '%s'. Please specify one of ['aws', 'gcp', 'azure', 'onprem'].", *cloudProvider)
		return false
	}
}

// func isValidInstanceOrServer() bool {
func isValidFlag() bool {
	if (*instanceArg == "" && !*serverArg && !*autobotsArg) ||
		(*instanceArg != "" && (*serverArg || *autobotsArg)) {
		logrus.Error("Flags should be provided in a mutually exclusive manner: 'instance', 'server/os', or 'autobots'.")
		return false
	}
	return true
}

func main() {
	kingpin.Version(version)

	kingpin.Parse()
	logrus.Infof("Parsed Flags: cloudProvider=%s, instanceArg=%s, serverArg=%v", *cloudProvider, *instanceArg, *serverArg)

	if !validateFlags() {
		kingpin.Usage()
		os.Exit(1)
	}

	// Setting up the logger
	helpers.SetupLogger(*debug)

	exeDir := schema.GetExecutableDir()
	schema.YamlCmdConfigFilePath = schema.GetConfigPath(exeDir, schema.DefaultCmdConfigYAMLPath)
	// Validate the CmdConfig.yaml file
	if err := schema.ValidateCmdConfigYAML(schema.YamlCmdConfigFilePath); err != nil {
		logrus.Fatal("Error validating cmd_config.yaml:", err)
	}

	if *serverArg {
		if err := tools.HandleOsCommands(*cloudProvider, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling OS commands:", err)
		}
	}

	if *instanceArg != "" {
		if err := tools.HandleP4Commands(*instanceArg, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling P4 commands:", err)
		}
	}
	if *autobotsArg {
		if err := tools.HandleAutobotsScripts(*OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling autobots scripts:", err)
		}
	}
	logrus.Info("Command-runner completed.")
}
