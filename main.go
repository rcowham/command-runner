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
	//	autobotsArg             = kingpin.Flag("autobots", "Enable running autobots scripts").Short('a').Bool() // TODO [TEMP]  removed short tor safty
	debug                   = kingpin.Flag("debug", "Enable debug logging").Short('d').Bool()
	cloudProvider           = kingpin.Flag("cloud", "Cloud provider (aws, gcp, azure, or onprem)").Short('c').Default("onprem").String()
	instanceArg             = kingpin.Flag("instance", "SDP instance commands argument for the command-runner").Short('i').String()                        //TODO [TEMP] .Required()
	ProccessAllSDPinstances = kingpin.Flag("allSDP", "Run on all SDP instances commands argument for the command-runner").Default("false").Hidden().Bool() //TODO [TEMP] Hidden for safety
	serverArg               = kingpin.Flag("server", "OS commands argument for the command-runner").Short('s').Bool()                                      //TODO [TEMP].Required()
	autoCloudFlag           = kingpin.Flag("autocloud", "Auto detect cloud provider").Hidden().Bool()                                                      //TODO [TEMP] Hidden for safety
	autobotsArg             = kingpin.Flag("autobots", "Enable running autobots scripts").Hidden().Bool()                                                  //TODO [TEMP]  Hidden for safety
	MainLogFilePath         = kingpin.Flag("log", "Path to the write the log file").Short('l').Default(schema.MainLogFilePath).String()                    //[TEMP] .Required()
	OutputJSONFilePath      = kingpin.Flag("output", "Path to the output JSON file").Short('o').Default(schema.OutputJSONFilePath).String()
	MetricsConfigFile       = kingpin.Flag("mcfg", "Path to the metrics configuration file").Default(schema.MetricsConfigFile).Short('m').String()
	//CmdConfigYAMLPath       = kingpin.Flag("cmdcfg", "Path to the cmd_config.yaml file").Default(schema.DefaultCmdConfigYAMLPath).Short('y').String()
	DefaultCmdConfigYAMLPath = kingpin.Flag("cmdcfg", "Path to the cmd_config.yaml file").Default(schema.DefaultCmdConfigYAMLPath).Short('y').String()
	nodelOut                 = kingpin.Flag("nodel", "Delete json data after running [default: true]").Default("false").Bool()
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
func GetDefaultCmdConfigYAMLPath() string {
	return *DefaultCmdConfigYAMLPath
}

// Check for valid flags
func isValidFlag() bool {
	// If no flags are provided
	if !*debug && *cloudProvider == "onprem" && *instanceArg == "" && !*ProccessAllSDPinstances && !*serverArg && !*autoCloudFlag && !*autobotsArg && *MainLogFilePath == schema.MainLogFilePath && *OutputJSONFilePath == schema.OutputJSONFilePath && *MetricsConfigFile == schema.MetricsConfigFile && *DefaultCmdConfigYAMLPath == schema.DefaultCmdConfigYAMLPath && *nodelOut {
		logrus.Error("At least one valid flag must be provided.")
		return false
	}
	// Check if only one of 'instance' or 'server' flags is provided
	// TODO This is temporary as this project evolves
	if (*instanceArg == "" && *serverArg) || (*instanceArg != "" && !*serverArg) {
		logrus.Error("Both 'instance' and 'server' flags should be provided together.")
		return false
	}

	// If autobotsArg is set but no context (instance or server) is provided, return false
	if *autobotsArg && (*instanceArg == "" || !*serverArg) {
		logrus.Error("'autobots' requires both 'instance' and 'server' to be specified.")
		return false
	}

	// If autoCloudFlag is set but serverArg is not provided, return false
	// (This condition might be redundant now as the serverArg is always expected to be there with instanceArg)
	if *autoCloudFlag && !*serverArg {
		logrus.Error("'autocloud' requires 'server' to be specified.")
		return false
	}

	// If autoCloudFlag is set, cloudProvider should not be manually set (or it should be set to "onprem" by default)
	if *autoCloudFlag && *cloudProvider != "onprem" {
		logrus.Error("When using 'autocloud', the 'cloud' flag should not be manually set.")
		return false
	}
	// Ensure --log is provided when either --server or --instance is provided
	//TODO TEMP
	if (*serverArg || *instanceArg != "") && *MainLogFilePath == "" {
		logrus.Error("The 'log' flag is required when using 'server' or 'instance'.")
		return false
	}
	if !*serverArg && *instanceArg == "" && !*autobotsArg && !*autoCloudFlag && *cloudProvider == "onprem" && *MainLogFilePath == "" {
		logrus.Error("At least one valid flag must be provided.")
		return false
	}
	return true
}

func main() {

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(version.Print("command-runner")).Author("Will Kreitzmann")
	kingpin.CommandLine.Help = "Runs a configurable set of commands and collects and reports the results as JSON for server/system monitoring\n"
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	// Setting up the logger
	helpers.SetupLogger(*debug, *MainLogFilePath)
	//logrus.Infof("Parsed Flags: cloudProvider=%s, instanceArg=%s, serverArg=%v", *cloudProvider, *instanceArg, *serverArg)
	logrus.Infof("Parsed Flags: debug=%v, cloudProvider=%s, instanceArg=%s, serverArg=%v ...", *debug, *cloudProvider, *instanceArg, *serverArg)

	if !validateFlags() {
		kingpin.Usage()
		os.Exit(1)
	}

	tools.GetVars(*DefaultCmdConfigYAMLPath)

	//exeDir := schema.GetExecutableDir()                                             //TODO MOVE THIS
	//schema.YamlCmdConfigFilePath = schema.GetConfigPath(exeDir, *CmdConfigYAMLPath) //TODO FIX THIS
	// Validate the CmdConfig.yaml file
	if err := schema.ValidateCmdConfigYAML(*DefaultCmdConfigYAMLPath); err != nil {
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
		if *autobotsArg {
			logrus.Infof("Running P4 SDP autobots...")
			tools.HandleOSAutobotsScripts(*OutputJSONFilePath, "") //TODO Look at blank
		}
		//TODO this doesn't needd to happen ever time does it?
		tools.FindP4D()
		if tools.P4dInstalled {
			// Do something if p4d is installed
			if *ProccessAllSDPinstances {
				if err := tools.GetSDPInstances(*OutputJSONFilePath, *autobotsArg, true, *debug); err != nil {
					logrus.Fatal("Error handling SDP instances:", err)
				}
			} else { // *allSDPinstances is false
				if err := tools.GetSDPInstances(*OutputJSONFilePath, *autobotsArg, false, *debug); err != nil {
					logrus.Fatal("Error handling SDP instances:", err)
				}
			}
		}
		if tools.P4dRunning {
			// Do something if p4d is running
		}
	}

	// Lets party for --instance=
	if *instanceArg != "" {
		if err := tools.HandleSDPInstance(*OutputJSONFilePath, *instanceArg, *autobotsArg, *debug); err != nil {
			logrus.Fatal("Error handling P4 commands:", err)
		}
	}
	if err := tools.PushToDataPushGateway(*OutputJSONFilePath, *MetricsConfigFile); err != nil {
		logrus.Fatal("Error Pushing to Data Push Gateway", err) //TODO this is possible not the right message
	}

	if !*nodelOut {
		// Delete the DefaultOutputJSONPath file
		if err := os.Remove(*OutputJSONFilePath); err != nil {
			logrus.Errorf("Error deleting file %s: %v", *OutputJSONFilePath, err)
		} else {
			logrus.Infof("Successfully deleted file: %s", *OutputJSONFilePath)
		}
	}
	logrus.Info("Command-runner completed.")
}
