package cmd

import (
	"github.com/spf13/cobra"
)

// countPicturesCmd represents the countPictures command
var countPicturesCmd = &cobra.Command{
	Use:   "count",
	Short: "Count the number of pictures stored on the camera",
	Run: func(cmd *cobra.Command, args []string) {
		count, err := getSerialClient().CountPictures()
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		cmd.Printf("Number of pictures stored on the camera: %d\n", count)
	},
}

func init() {
	rootCmd.AddCommand(countPicturesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// countPicturesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// countPicturesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
