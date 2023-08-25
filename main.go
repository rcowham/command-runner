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
	// If both instanceArg and serverArg are specified, return false
	if *instanceArg != "" && *serverArg {
		logrus.Error("Flags 'instance' and 'server' should not be used together.")
		return false
	}

	// If neither instanceArg nor serverArg is provided and autobotsArg is not set, return false
	if *instanceArg == "" && !*serverArg && !*autobotsArg {
		logrus.Error("Either 'instance' or 'server' flag should be provided.")
		return false
	}

	// If autobotsArg is set and instanceArg is provided, return false
	if *instanceArg != "" && *autobotsArg {
		logrus.Error("'autobots' cannot be used with 'instance'.")
		return false
	}

	// If autobotsArg is set but serverArg is not provided, return false
	if !*serverArg && *autobotsArg {
		logrus.Error("'autobots' requires 'server' to be specified.")
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
		// Handle server logic here
		if err := tools.HandleOsCommands(*cloudProvider, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling OS commands:", err)
		}
		tools.FindP4D()
		if tools.P4dInstalled {
			// Do something if p4d is installed
			if err := tools.GetSDPInstances(*OutputJSONFilePath, *autobotsArg); err != nil {
				logrus.Fatal("Error handling SDP instances:", err)
			}
		}
		if tools.P4dRunning {
			// Do something if p4d is running
		}
	}

	if *instanceArg != "" {
		if err := tools.HandleP4Commands(*instanceArg, *OutputJSONFilePath); err != nil {
			logrus.Fatal("Error handling P4 commands:", err)
		}
	}
	/*
		if *autobotsArg {
			if err := tools.HandleAutobotsScripts(*OutputJSONFilePath, ""); err != nil {
				logrus.Fatal("Error handling autobots scripts:", err)
			}
		}
	*/
	if err := tools.PushToDataPushGateway(*OutputJSONFilePath, "/p4/common/config/.push_metrics.cfg"); err != nil {
		logrus.Fatal("Error handling autobots scripts:", err)
	}

	logrus.Info("Command-runner completed.")
}
