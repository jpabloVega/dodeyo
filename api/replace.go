package api

import (
	"strings"

	vocabtrie "github.com/jpabloVega/dokodemoYomou/vocab_trie"
)

func ReplaceKanji(chapter Chapter) (Chapter, error) {
	var replacedCont []string

	t := vocabtrie.NewTrie()
	err := t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
	if err != nil {
		return Chapter{}, err
	}

	for _, line := range chapter.Contents {
		if line == "" {
			continue
		}

		runes := []rune(line)
		var resultStr strings.Builder
		start := 0
		for start < len(runes) {
			var skips int
			var foundWord vocabtrie.Word

			for j := len(runes); j >= start; j-- {
				var wasFound bool
				currWord := string(runes[start:j])
				foundWord, wasFound = t.Search(currWord)
				if wasFound {
					skips = len(runes[start:j])
					break
				}
			}
			if foundWord.Kana != "" {
				resultStr.WriteString(foundWord.Kana)
				start += skips
			} else {
				if foundWord.Kanji == "" {
					resultStr.WriteRune(runes[start])
					start += 1
				}
				resultStr.WriteString(foundWord.Kanji)
				start += skips
			}
		}
		replacedCont = append(replacedCont, resultStr.String())
	}
	return Chapter{
		StatusCode: chapter.StatusCode,
		Title:      chapter.Title,
		Contents:   replacedCont,
	}, nil
}
