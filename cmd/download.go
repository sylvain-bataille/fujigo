/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download pictures from the camera",
	Long:  `Usage: fujigo download [image_number]`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("download called")
		imageNumber := 0
		_, err := fmt.Sscanf(args[0], "%d", &imageNumber)
		if err != nil || imageNumber < 1 || imageNumber > 65535 {
			cmd.PrintErr("Invalid image number. It must be between 1 and 65535.\n")
			return
		}
		data, err := getSerialClient().DownloadPicture(imageNumber)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		if len(data) == 0 {
			cmd.Println("No data received for the picture.")
			return
		}
		// Print data bytes
		cmd.Printf("Data bytes: %X\n", data)
		// Save to file
		filename := fmt.Sprintf("image_%d.jpg", imageNumber)
		err = saveToFile(filename, data)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		cmd.Printf("Picture %d saved to %s\n", imageNumber, filename)
	},
}

func saveToFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
