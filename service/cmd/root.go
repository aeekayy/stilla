/*
Copyright © 2023 Farye Nwede <farye@aeekay.com>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"github.com/aeekayy/stilla/service/pkg/service"
)

var (
	configFile string
	cpuProfile     bool
	memProfile     bool
	cpuProfileFile string
	memProfileFile string
)


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stilla",
	Short: "Stilla Configuration Management",
	Long: `A configuration Management service. This service makes
configuration available for scripts, modules, and services.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		onStopProfiling = profilingInit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := runService()
		// On the most outside function we only log error
		if err != nil {
			fmt.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer stopProfiling()
	_ = rootCmd.Execute()
}

// init is called before main
func init() {
	rootCmd.Flags().StringVar(&configFile, "config", "", "Configuration file for Stilla")

	// Profiling cli flags
	rootCmd.PersistentFlags().BoolVar(&cpuProfile, "cpu-profile", false, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&memProfile, "mem-profile", false, "write memory profile to file")

	rootCmd.PersistentFlags().StringVar(&cpuProfileFile, "cpu-profile-file", "cpu.prof", "write cpu profile to file")
	rootCmd.PersistentFlags().StringVar(&memProfileFile, "mem-profile-file", "mem.prof", "write memory profile to file")
}

// runService run the service
func runService() (err error) {
	// Print the config
	if cpuProfile || memProfile {
		fmt.Printf(
			"Config: {\n\tCPUProfile: %t\n\tCPUProfileFile: %s\n\tMEMProfile: %t\n\tMEMProfileFile: %s\n}\n",
			cpuProfile, cpuProfileFile, memProfile, memProfileFile)
	}

	// Create new app instance
	svc := service.NewService(configFile)

	return svc.Start()
}