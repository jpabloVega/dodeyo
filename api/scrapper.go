package api

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/schollz/progressbar/v3"
)

type NovelType int

const (
	Serialized NovelType = iota
	Short
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0"

type UrlInfo struct {
	Site       string
	Url        string
	StartIndex int
}

type Chapter struct {
	StatusCode int
	Title      string
	Contents   []string
}

type Credits struct {
	Author string
	Title  string
	URL    string
	Type   NovelType
	Body   []string
}

func (c *Client) GetSingleAddress(address string) (Chapter, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return Chapter{}, err
	}
	req.Header.Set("User-Agent", UserAgent)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return Chapter{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return Chapter{StatusCode: res.StatusCode}, nil
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Chapter{}, err
	}
	chapterData, err := GetChapterData(string(data))
	if err != nil {
		return Chapter{}, err
	}
	return chapterData, nil
}

func GetSiteInfo(baseURL string) (UrlInfo, error) {
	urlS, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return UrlInfo{}, err
	}
	if urlS.Host != "ncode.syosetu.com" {
		return UrlInfo{}, errors.New("The url must come from ncode.syosetu.com")
	}
	urlParts := strings.Split(urlS.Path, "/")
	startIndex := 1
	if urlParts[2] != "" {
		index, err := strconv.Atoi(urlParts[2])
		if err != nil {
			return UrlInfo{}, err
		}
		if index > 0 {
			startIndex = index
		}
	}
	fullURL := url.URL{
		Scheme: urlS.Scheme,
		Host:   urlS.Host,
		Path:   urlParts[1],
	}
	return UrlInfo{
		Site:       urlParts[0],
		Url:        fullURL.String(),
		StartIndex: startIndex,
	}, nil
}

func (c *Client) GetChapterAuthor(url string) (Credits, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Credits{}, err
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return Credits{}, err
	}
	defer res.Body.Close()
	if res.StatusCode > 400 {
		return Credits{}, err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Credits{}, err
	}
	titleFind := regexp.MustCompile(`<meta property="og:title" content="\s*(.*?)\s*">`)
	title := titleFind.FindAllStringSubmatch(string(data), 1)
	authorFind := regexp.MustCompile(`<meta name="twitter:creator" content="\s*(.*?)\s*">`)
	author := authorFind.FindAllStringSubmatch(string(data), 1)

	var novelType NovelType
	var body []string
	if strings.Contains(string(data), `<div class="p-novel__body">`) {
		novelType = Short
		body = getShortBody(string(data))
	} else {
		novelType = Serialized
		body = getSummary(string(data))
	}
	return Credits{
		Title:  title[0][1],
		Author: author[0][1],
		URL:    url,
		Type:   novelType,
		Body:   body,
	}, nil
}

func getSummary(data string) []string {
	removedTop := strings.Split(data, `<div id="novel_ex" class="p-novel__summary">`)
	summBody := strings.Split(removedTop[1], `<div class="c-ad c-ad--layout-horizontal">`)
	lines := strings.Split(summBody[0], `<br />`)
	var finalStrs []string
	pText := bluemonday.StripTagsPolicy()
	for _, line := range lines {
		cleanLine := pText.Sanitize(line)
		if cleanLine == "" {
			continue
		}
		finalStrs = append(finalStrs, cleanLine)
	}
	return finalStrs
}

func getShortBody(data string) []string {
	removedTop := strings.Split(data, `<div class="js-novel-text p-novel__text">`)
	summBody := strings.Split(removedTop[1], `</div>`)
	lines := strings.Split(summBody[0], `</p>`)
	var finalStrs []string
	pText := bluemonday.StripTagsPolicy()
	for _, line := range lines {
		cleanLine := pText.Sanitize(line)
		if cleanLine == "" {
			continue
		}
		finalStrs = append(finalStrs, cleanLine)
	}
	return finalStrs
}

func GetChapterData(data string) (Chapter, error) {
	var lines []string
	titleFind := regexp.MustCompile(`<h1 class="p-novel__title p-novel__title--rensai">\s*(.*?)\s*</h1>`)
	title := titleFind.FindAllStringSubmatch(data, 1)
	topCrop := strings.Split(data, `<div class="js-novel-text p-novel__text">`)
	if len(topCrop) < 2 {
		return Chapter{}, errors.New("Couldnt find body, possible bad request")
	}
	if len(title[0]) < 2 {
		return Chapter{}, errors.New("Couldnt find title")
	}
	pText := bluemonday.StripTagsPolicy()
	botCrop := strings.Split(topCrop[1], "</div>")
	filteredBody := strings.Split(botCrop[0], "</p>")
	for _, line := range filteredBody {
		cleanText := pText.Sanitize(line)
		clean := strings.TrimSpace(cleanText)
		lines = append(lines, clean)
	}
	cleanTitle := pText.Sanitize(title[0][1])
	return Chapter{
		StatusCode: 200,
		Title:      cleanTitle,
		Contents:   lines,
	}, nil

}

func (c *Client) GetFromRange(urlData UrlInfo, chapterAmount int) ([]Chapter, error) {
	downloadBar := progressbar.Default(int64(chapterAmount), "Downloading chapters...")
	var scrappedChapters []Chapter
	for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
		chapter, err := c.GetSingleAddress(fullURL)
		if chapter.StatusCode == http.StatusNotFound {
			break
		}
		if err != nil {
			return []Chapter{}, err
		}
		time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
		downloadBar.Add(1)
		scrappedChapters = append(scrappedChapters, chapter)
	}
	return scrappedChapters, nil
}
