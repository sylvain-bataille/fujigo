package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// listPortsCmd represents the listPorts command
var listPortsCmd = &cobra.Command{
	Use:   "list-ports",
	Short: "List available serial ports",
	Run: func(cmd *cobra.Command, args []string) {
		ports, err := fuji.ListPorts()
		if err != nil {
			cmd.PrintErrln("Error listing ports:", err)
			return
		}
		if len(ports) == 0 {
			cmd.Println("No serial ports found")
			return
		}
		cmd.Println("Available serial ports:")
		for _, port := range ports {
			cmd.Println(" -", port)
		}
	},
}

func init() {
	rootCmd.AddCommand(listPortsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listPortsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listPortsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
