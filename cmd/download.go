package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download [image_number|all]",
	Short: "Download pictures from the camera",
	Long: `Downloads pictures from the camera by image number or all images.
	The first image is number 1. 
	Use count command to see how many images are available.
	Example to download image number 3: fujigo download 3
	Example to download all images: fujigo download all`,
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
		saveAllImagesToFile(selectedImages, serialClient, cmd)
		deleteAfter, _ := cmd.Flags().GetBool("delete")
		if deleteAfter {
			deleteSelectedImages(selectedImages, serialClient, cmd)
		}
	},
}

// collectImageIndices parses the command line arguments to determine which images to download.
// Parameters are the command line arguments and the total count of images available.
// It returns a slice of all image indices in case of "all" argument, or a slice with a single image index.
func collectImageIndices(args []string, count int) ([]int, error) {
	imgNumberArg := 0
	selectedImages := []int{}
	if len(args) != 1 {
		return nil, fmt.Errorf("Please provide exactly one argument: image number or 'all'.")
	}
	if args[0] == "all" {
		for i := 1; i <= count; i++ {
			selectedImages = append(selectedImages, i)
		}
	} else {
		_, err := fmt.Sscanf(args[0], "%d", &imgNumberArg)
		if err != nil || imgNumberArg < 1 || imgNumberArg > 65535 {
			return nil, fmt.Errorf("Invalid image number. It must be between 1 and 65535.")
		}
		selectedImages = append(selectedImages, imgNumberArg)
	}
	return selectedImages, nil
}

// saveAllImagesToFile downloads and saves all selected images to files.
func saveAllImagesToFile(selectedImages []int, serialClient *fuji.SerialClient, cmd *cobra.Command) {
	for _, imgNum := range selectedImages {
		cmd.Printf("Downloading picture %d...\n", imgNum)
		start := time.Now()
		data, err := serialClient.DownloadPicture(imgNum)
		end := time.Now()
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		if len(data) == 0 {
			cmd.Println("No data received for the picture.")
			return
		}
		filename := fmt.Sprintf("image_%d.jpg", imgNum)
		err = saveToFile(filename, data)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		duration := end.Sub(start)
		cmd.Printf("Downloaded %d bytes in %.2f seconds (%.2f KB/s)\n", len(data), duration.Seconds(), float64(len(data))/duration.Seconds()/1024)
		cmd.Printf("Picture %d saved to %s\n", imgNum, filename)
	}
}

// saveToFile saves the given data to a file with the specified filename.
// file is located in the current working directory.
func saveToFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().Bool("delete", false, "Delete pictures from the camera after downloading")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
