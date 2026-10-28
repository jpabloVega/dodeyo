package cmd

import (
	"fmt"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/jpabloVega/dokodemoYomou/database"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(historyCmd)
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Download history",
	Long:  "Check your download history",
	RunE:  GetHistory,
}

func GetHistory(_ *cobra.Command, _ []string) error {
	db, err := database.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	books, err := database.GetAllBooks(db)
	if err != nil {
		return err
	}

	fmt.Println("Download history:")
	for _, book := range books {
		fmt.Printf(" Name: %s\n", book.Name)
		fmt.Printf(" Link: %s\n", book.Link)
		if book.NovelType == api.Serialized {
			fmt.Printf(" Resume download from chapter: %d\n", book.LastDownload)
		} else {
			fmt.Println(" One shot")
		}
		fmt.Printf(" Downloaded at : %s\n", book.DownloadedAt)
		fmt.Println()
	}
	return nil
}
