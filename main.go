/*
 */
package main

import (
	"command-runner/helpers"
	"command-runner/schema"
	"command-runner/tools"

	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/perforce/p4prometheus/version"
	"github.com/sirupsen/logrus"
)

var (
	cloudProvider      = kingpin.Flag("cloud", "Cloud provider (aws, gcp, azure, or onprem)").Short('c').Default("onprem").String()
	debug              = kingpin.Flag("debug", "Enable debug logging").Short('d').Bool()
	OutputJSONFilePath = kingpin.Flag("output", "Path to the output JSON file").Short('o').Default(schema.OutputJSONFilePath).String()
	instanceArg        = kingpin.Flag("instance", "SDP instance commands argument for the command-runner").Short('i').String()
	serverArg          = kingpin.Flag("server", "OS commands argument for the command-runner").Short('s').Bool()
	autoCloudFlag      = kingpin.Flag("autocloud", "Auto detect cloud provider").Bool()
	autobotsArg        = kingpin.Flag("autobots", "Enable running autobots scripts").Short('a').Bool()
	MetricsConfigFile  = kingpin.Flag("mcfg", "Path to the metrics configuration file").Default(schema.MetricsConfigFile).Short('m').String()
	CmdConfigYAMLPath  = kingpin.Flag("cmdcfg", "Path to the cmd_config.yaml file").Default(schema.DefaultCmdConfigYAMLPath).Short('y').String()
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

// Check for valid flags
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

	// If autoCloudFlag is set but serverArg is not provided, return false
	if *autoCloudFlag && !*serverArg {
		logrus.Error("'autocloud' requires 'server' to be specified.")
		return false
	}

	// If autoCloudFlag is set, cloudProvider should not be manually set (or it should be set to "onprem" by default)
	if *autoCloudFlag && *cloudProvider != "onprem" {
		logrus.Error("When using 'autocloud', the 'cloud' flag should not be manually set.")
		return false
	}

	return true
}

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version.Print("command-runner")).Author("Will Kreitzmann")
	kingpin.CommandLine.Help = "Runs a configurable set of commands and collects and reports the results as JSON for server/system monitoring\n"
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logrus.Infof("Parsed Flags: cloudProvider=%s, instanceArg=%s, serverArg=%v", *cloudProvider, *instanceArg, *serverArg)

	if !validateFlags() {
		kingpin.Usage()
		os.Exit(1)
	}

	// Setting up the logger
	helpers.SetupLogger(*debug)

	exeDir := schema.GetExecutableDir()
	schema.YamlCmdConfigFilePath = schema.GetConfigPath(exeDir, *CmdConfigYAMLPath)
	// Validate the CmdConfig.yaml file
	if err := schema.ValidateCmdConfigYAML(schema.YamlCmdConfigFilePath); err != nil {
		logrus.Fatal("Error validating cmd_config.yaml:", err)
	}

	if *serverArg {
		// If autoCloudFlag is enabled, detect the cloud provider
		if *autoCloudFlag {
			detectedCloudProvider, err := tools.DetectCloudProvider()
			// Assuming you have a function DetectCloudProvider in your tools package
			if err != nil {
				logrus.Fatal("Error detecting cloud provider:", err)
			}

			// Update cloudProvider variable with the detected value
			*cloudProvider = detectedCloudProvider

			// Update the metrics config with the detected cloud provider
			if err := schema.UpdateMetricsConfig(detectedCloudProvider); err != nil {
				logrus.Fatal("Error updating metrics configuration:", err)
			}
		}
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
	if err := tools.PushToDataPushGateway(*OutputJSONFilePath, *MetricsConfigFile); err != nil {
		logrus.Fatal("Error handling autobots scripts:", err)
	}

	// Delete the DefaultOutputJSONPath file
	if err := os.Remove(*OutputJSONFilePath); err != nil {
		logrus.Errorf("Error deleting file %s: %v", *OutputJSONFilePath, err)
	} else {
		logrus.Infof("Successfully deleted file: %s", *OutputJSONFilePath)
	}

	logrus.Info("Command-runner completed.")
}
