package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// previewCmd represents the preview command
var previewCmd = &cobra.Command{
	Use:   "preview [image_number|all]",
	Short: "Preview thumbnails from the camera",
	Long: `Downloads thumbnails from the camera by image number or all images.
	The first image is number 1. 
	Use count command to see how many images are available.
	Thumbnails are smaller and faster to download than full images.
	Example to preview image number 3: fujigo preview 3
	Example to preview all images: fujigo preview all`,
	Args: cobra.MatchAll(cobra.ExactArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		serialClient := getSerialClient()
		count, error := serialClient.CountPictures()
		if error != nil {
			cmd.PrintErr(error)
			return
		}
		if count == 0 {
			cmd.Println("No pictures found on the camera.")
			return
		}
		selectedImages, err := collectImageIndices(args, count)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		saveAllThumbnailsToFile(selectedImages, serialClient, cmd)
	},
}

// saveAllThumbnailsToFile downloads and saves all selected thumbnails to files.
func saveAllThumbnailsToFile(selectedImages []int, serialClient *fuji.SerialClient, cmd *cobra.Command) {
	for _, imgNum := range selectedImages {
		cmd.Printf("Downloading thumbnail %d...\n", imgNum)
		start := time.Now()
		data, err := serialClient.DownloadThumbnail(imgNum)
		end := time.Now()
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		if len(data) == 0 {
			cmd.Println("No data received for the thumbnail.")
			return
		}
		filename := fmt.Sprintf("thumbnail_%d.jpg", imgNum)
		err = saveToFile(filename, data)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		duration := end.Sub(start)
		cmd.Printf("Downloaded %d bytes in %.2f seconds (%.2f KB/s)\n", len(data), duration.Seconds(), float64(len(data))/duration.Seconds()/1024)
		cmd.Printf("Thumbnail %d saved to %s\n", imgNum, filename)
	}
}

func init() {
	rootCmd.AddCommand(previewCmd)
}
