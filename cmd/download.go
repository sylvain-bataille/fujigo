package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download pictures from the camera",
	Long:  `Usage: fujigo download [image_number|all] - Downloads pictures from the camera by image number or all images.`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		imageNumber := 0
		if len(args) != 1 {
			cmd.PrintErr("Please provide exactly one argument: image number or 'all'.\n")
			return
		}
		photos := []int{}
		count, error := getSerialClient().CountPictures()
		if error != nil {
			cmd.PrintErr(error)
			return
		}
		if count == 0 {
			cmd.Println("No pictures found on the camera.")
			return
		}
		if args[0] == "all" {
			for i := 1; i <= count; i++ {
				photos = append(photos, i)
			}
		} else {
			_, err := fmt.Sscanf(args[0], "%d", &imageNumber)
			if err != nil || imageNumber < 1 || imageNumber > 65535 {
				cmd.PrintErr("Invalid image number. It must be between 1 and 65535.\n")
				return
			}
			photos = append(photos, imageNumber)
		}
		serialClient := getSerialClient()
		for _, imgNum := range photos {
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
			cmd.Printf("Downloaded %d bytes in %v (%.2f KB/s)\n", len(data), duration, float64(len(data))/duration.Seconds()/1024)
			cmd.Printf("Picture %d saved to %s\n", imgNum, filename)
		}
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
