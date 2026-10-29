package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jpabloVega/dodeyo/api"
	"github.com/jpabloVega/dodeyo/email"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(cnvrtCmd)

	cnvrtCmd.Flags().IntVar(&bodySize, "bodysize", 14, "Size for the font of the body (only for pdf)")
	cnvrtCmd.Flags().IntVar(&titleSize, "titlesize", 16, "Size for the font of the titles (only for pdf)")
	cnvrtCmd.Flags().StringVarP(&destinationPath, "saveto", "s", "", "Path to the folder output file's are stored")
	cnvrtCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the output file, defaults to novel title")
	cnvrtCmd.Flags().BoolVarP(&vertical, "vertical", "v", false, "Export text in vertical(only availible to html and epub)")
	cnvrtCmd.Flags().BoolVar(&mailToKindle, "send", false, "Mail to kindle")
	cnvrtCmd.Flags().BoolVar(&replaceKanji, "furigana", false, "Add furigana")
	cnvrtCmd.Flags().BoolVar(&oPDF, "pdf", false, "Output in PDF")
	cnvrtCmd.Flags().BoolVar(&oHTML, "html", false, "Output in HTML")
	cnvrtCmd.Flags().BoolVar(&oEPub, "epub", false, "Output in EPUB")
	cnvrtCmd.MarkFlagsOneRequired("pdf", "html", "epub")
	cnvrtCmd.MarkFlagsMutuallyExclusive("pdf", "html", "epub")
}

var cnvrtCmd = &cobra.Command{
	Use:   "convert [filepath .txt]",
	Short: "Convert txt files to eReader format",
	Long:  "Convert txt files into a valid format for eReader",
	Args:  cobra.ExactArgs(1),
	RunE:  convertCmd,
}

func convertCmd(_ *cobra.Command, args []string) error {
	var docs []string
	for _, path := range args {
		filePath, err := convertFile(path)
		if err != nil {
			return err
		}
		docs = append(docs, filePath)
	}

	if mailToKindle {
		err := email.MailToKindle(docs)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertFile(filePath string) (string, error) {
	var err error
	if destinationPath == "" {
		destinationPath = viper.GetString("defaultpath")
		if destinationPath == "" {
			destinationPath, err = setDestinationPath()
			if err != nil {
				return "", err
			}
		}
	}

	title, err := checkValidFile(filePath)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	chaptersData, bookCredits := convertToChapter(string(content), title)

	if replaceKanji {
		furiChap := []api.Chapter{}
		for _, chap := range chaptersData {
			repChap, err := api.ReplaceKanji(chap)
			if err != nil {
				return "", err
			}
			furiChap = append(furiChap, repChap)
		}
		chaptersData = furiChap
	}

	var path string

	if oPDF {
		path, err = GetPDF(chaptersData, bookCredits)
	}
	if oHTML {
		path, err = GetHTML(chaptersData, bookCredits)
	}
	if oEPub {
		path, err = GetEPub(chaptersData, bookCredits)
	}
	if err != nil {
		return "", err
	}

	return path, nil

}

func checkValidFile(filePath string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("Invalid file path")
		}
		return "", fmt.Errorf("Error accessing path: %v\n", err)
	}

	if info.IsDir() {
		return "", errors.New("No file in path")
	}

	fileExt := filepath.Ext(filePath)
	if fileExt != ".txt" {
		return "", errors.New("Invalid file extension")
	}

	rawTitle := filepath.Base(filePath)
	title := strings.TrimSuffix(rawTitle, fileExt)
	return title, nil
}

func convertToChapter(contents, title string) ([]api.Chapter, api.Credits) {
	var chapters []api.Chapter
	var credits api.Credits
	credits.Title = title
	scanner := bufio.NewScanner(strings.NewReader(contents))
	isTitle := true
	creditOrder := 0
	var currChap api.Chapter

	for scanner.Scan() {
		if scanner.Text() == "Credits" {
			creditOrder++
			continue
		}
		if creditOrder > 0 {
			if creditOrder == 1 {
				credits.Author = scanner.Text()
				creditOrder++
				continue
			}
			if creditOrder == 2 {
				credits.URL = scanner.Text()
				creditOrder++
				continue
			}
			if creditOrder > 2 {
				break
			}
		}
		if scanner.Text() == "" {
			isTitle = true
			chapters = append(chapters, currChap)
			currChap = api.Chapter{}
			continue
		}
		if isTitle {
			currChap.Title = scanner.Text()
			isTitle = false
			continue
		}
		currChap.Contents = append(currChap.Contents, scanner.Text())
	}
	return chapters, credits
}
