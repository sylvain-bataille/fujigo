package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fujigo",
	Short: "Fujigo is a CLI tool to interact with old Fujifilm cameras over serial",
	Long: `This application is a tool to interact with old Fujifilm cameras over serial.
	It currently supports to list serial ports, print camera info, count and download pictures.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fujigo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("device", "d", "/dev/ttyUSB0", "the serial device to use (e.g., /dev/ttyUSB0 or COM3)")

}

func getDevice() string {
	device, _ := rootCmd.Flags().GetString("device")
	return device
}

func isVerbose() bool {
	verbose, _ := rootCmd.Flags().GetBool("verbose")
	return verbose
}

func getSerialClient() *fuji.SerialClient {
	return fuji.NewSerialClient(isVerbose(), getDevice(), 9600)
}
