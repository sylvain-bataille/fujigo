package cmd

import (
	"slices"

	"github.com/spf13/cobra"
	"github.com/sylvain-bataille/fujigo/fuji"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [picture_number|all]",
	Short: "Delete a picture by its number or all pictures",
	Long: `Delete a picture by its number or all pictures.
	If 'all' is specified, all pictures will be deleted.`,
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
		deleteSelectedImages(selectedImages, serialClient, cmd)
	},
}

// deleteSelectedImages deletes all images whose indices are in the selectedImages slice.
func deleteSelectedImages(selectedImages []int, serialClient *fuji.SerialClient, cmd *cobra.Command) {
	deletedImages := 0
	slices.Sort(selectedImages)
	selectedImages = slices.Compact(selectedImages)
	for _, imgNumber := range selectedImages {
		cmd.Printf("Deleting image %d...\n", imgNumber)
		err := serialClient.DeletePicture(imgNumber - deletedImages)
		if err != nil {
			cmd.Printf("Error deleting image %d: %v\n", imgNumber, err)
			continue
		}
		cmd.Printf("Image %d deleted successfully.\n", imgNumber)
		deletedImages++
	}
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
