package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fujigo",
	Short: "Fujigo is a CLI tool to interact with old Fujifilm cameras over serial",
	Long: `This application is a tool to interact with old Fujifilm cameras over serial.
	It currently supports listing available serial ports and getting information about the connected camera model.
	In the future, extracting images and other functionalities may be added.
	With the difficulty to read some SmartMedia cards on modern computers, this tool aims to provide an alternative way to access images stored on these cameras.
	This tool is based on this protocol analysis:  https://christian1.tripod.com/FujiMX.html, and older linux tool Fujiplay.`,
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
	rootCmd.Flags().BoolP("verbose", "v", false, "verbose output ... to be implemented")
	rootCmd.Flags().BoolP("device", "d", false, "the serial device to use (e.g., /dev/ttyUSB0 or COM3) ... to be implemented")

}
