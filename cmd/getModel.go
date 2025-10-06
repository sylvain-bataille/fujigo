package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// getModelCmd represents the get-model command
var getModelCmd = &cobra.Command{
	Use:   "get-model",
	Short: "Get information about the connected camera model",
	Run: func(cmd *cobra.Command, args []string) {
		fuji.GetModel()
	},
}

func init() {
	rootCmd.AddCommand(getModelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getModelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getModelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
