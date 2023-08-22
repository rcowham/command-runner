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
	instanceArg        = kingpin.Flag("instance", "Instance argument for the command-runner").Short('i').String()
	serverArg          = kingpin.Flag("server", "Server argument for the command-runner").Short('s').Bool()
	version            = "development"
)

func validateFlags() bool {
	return isValidProvider() && isValidInstanceOrServer()
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

func isValidInstanceOrServer() bool {
	if (*instanceArg == "" && !*serverArg) || (*instanceArg != "" && *serverArg) {
		logrus.Error("Either the 'instance' flag or the 'server' flag should be provided, but not both.")
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
		if err := tools.HandleServerCommands(*cloudProvider, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling server commands:", err)
		}
	}

	if *instanceArg != "" {
		if err := tools.HandleInstanceCommands(*instanceArg, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling instance commands:", err)
		}
	}

	logrus.Info("Command-runner completed.")
}
