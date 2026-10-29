package cmd

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jpabloVega/dodeyo/api"
	vocabtrie "github.com/jpabloVega/dodeyo/vocab_trie"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show details")
	scanCmd.Flags().IntVarP(&chapterAmount, "amount", "a", 10, "Download this amount of chapters")
	scanCmd.Flags().IntVar(&startFrom, "start", 0, "Start from chapter")

}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Get the urls",
	Long:  "Implement me",
	RunE:  ScanBook,
}

func ScanBook(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("No argument given")
	}
	c := api.NewClient(5 * time.Second)
	for _, arg := range args {
		err := ScanBookCommand(arg, &c)
		if err != nil {
			return err
		}
	}
	return nil
}

type wordsSorted struct {
	n1    []string
	n2    []string
	n3    []string
	n4    []string
	n5    []string
	other []string
}

func ScanBookCommand(address string, c *api.Client) error {
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}

	chapters, err := c.GetFromRange(urlData, chapterAmount)
	if err != nil {
		return err
	}

	var okChapters []api.Chapter

	for _, chap := range chapters {
		if chap.StatusCode != 200 {
			fmt.Printf("Error getting chapter: %d\n", chap.StatusCode)
			continue
		}
		okChapters = append(okChapters, chap)
	}

	chaptersData := okChapters

	if len(chaptersData) == 0 {
		return errors.New("No chapters to scan")
	}

	t := vocabtrie.NewTrie()
	dictReader, err := AssetsFS.Open("assets/dictionary.json")
	if err != nil {
		return err
	}
	defer dictReader.Close()

	err = t.LoadFromReader(dictReader)
	if err != nil {
		return err
	}

	resultsChan := make(chan wordsSorted, len(chaptersData))
	var wg sync.WaitGroup

	for _, chapter := range chaptersData {
		for _, line := range chapter.Contents {
			wg.Add(1)
			go func(currLine string) {
				defer wg.Done()

				if currLine == "" {
					resultsChan <- wordsSorted{}
					return
				}
				splitLine := []rune(currLine)
				start := 0
				end := len(splitLine)
				lineCount := wordsSorted{}
				for start < end {
					currWord := string(splitLine[start:end])
					foundWord, found := t.Search(currWord)
					if found {
						wordLen := end - start
						start += wordLen
						end = len(splitLine)
						switch foundWord.Level {
						case "n1":
							lineCount.n1 = append(lineCount.n1, currWord)
						case "n2":
							lineCount.n2 = append(lineCount.n2, currWord)
						case "n3":
							lineCount.n3 = append(lineCount.n3, currWord)
						case "n4":
							lineCount.n4 = append(lineCount.n4, currWord)
						case "n5":
							lineCount.n5 = append(lineCount.n5, currWord)
						default:
							lineCount.other = append(lineCount.other, currWord)
						}
						continue
					}
					end--
					if end == start {
						start++
						end = len(splitLine)
					}
				}
				resultsChan <- lineCount
			}(line)
		}
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	nCount := make(map[string][]string)

	for results := range resultsChan {
		nCount["n1"] = append(nCount["n1"], results.n1...)
		nCount["n2"] = append(nCount["n2"], results.n2...)
		nCount["n3"] = append(nCount["n3"], results.n3...)
		nCount["n4"] = append(nCount["n4"], results.n4...)
		nCount["n5"] = append(nCount["n5"], results.n5...)
		nCount["other"] = append(nCount["other"], results.other...)
	}

	if verbose {
		fmt.Printf("N1: %v\n", nCount["n1"])
		fmt.Println("")
		fmt.Printf("N2: %v\n", nCount["n2"])
		fmt.Println("")
		fmt.Printf("N3: %v\n", nCount["n3"])
		fmt.Println("")
		fmt.Printf("N4: %v\n", nCount["n4"])
		fmt.Println("")
		fmt.Printf("N5: %v\n", nCount["n5"])
		fmt.Println("")
		fmt.Printf("Other: %v\n", nCount["other"])
		fmt.Println("")
	}
	fmt.Println("Words from every level")
	fmt.Printf("N1: %d\n", len(nCount["n1"]))
	fmt.Printf("N2: %d\n", len(nCount["n2"]))
	fmt.Printf("N3: %d\n", len(nCount["n3"]))
	fmt.Printf("N4: %d\n", len(nCount["n4"]))
	fmt.Printf("N5: %d\n", len(nCount["n5"]))
	fmt.Printf("other: %d\n", len(nCount["other"]))
	return nil
}
