package cmd

import (
	"fmt"
	"os"

	"github.com/jpabloVega/dodeyo/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Get the urls",
	Long:  "Implement me",
	RunE:  helloTest,
}

func helloTest(_ *cobra.Command, args []string) error {
	//fmt.Println("Not implemented")
	err := testReplace()
	if err != nil {
		return err
	}
	return nil
}

func testReplace() error {
	content, err := os.ReadFile("api/test.txt")
	if err != nil {
		return err
	}

	chapter, err := api.GetChapterData(string(content))
	if err != nil {
		return err
	}

	kanaChapter, err := api.ReplaceKanji(chapter)
	if err != nil {
		return err
	}

	fmt.Println(kanaChapter.Title)
	for i, line := range kanaChapter.Contents {
		fmt.Println(chapter.Contents[i])
		fmt.Println(line)
		fmt.Println()

	}
	return nil
}
