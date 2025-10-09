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
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbose output (add multiple times for more verbosity : v or vv)")
	rootCmd.PersistentFlags().StringP("device", "d", "/dev/ttyUSB0", "the serial device to use (e.g., /dev/ttyUSB0 or COM3)")

}

func getDevice() string {
	device, _ := rootCmd.Flags().GetString("device")
	return device
}

func getVerboseLvl() int {
	lvl, _ := rootCmd.Flags().GetCount("verbose")
	return lvl
}

func getSerialClient() *fuji.SerialClient {
	return fuji.NewSerialClient(getVerboseLvl(), getDevice(), 9600)
}
