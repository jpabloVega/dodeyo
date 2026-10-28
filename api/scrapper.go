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

	"github.com/schollz/progressbar/v3"
)

type NovelType int

const (
	Serialized NovelType = iota
	Short
)

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

func (c *Client) GetSingleAddress(address, userAgent string) (Chapter, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return Chapter{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	//req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	//req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	//req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Upgrade-Insecure-Requests", "1")

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
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
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
	reg := regexp.MustCompile(`<\s*(.*?)\s*>`)
	for _, line := range lines {
		line = reg.ReplaceAllString(line, "")
		if line == "" {
			continue
		}
		finalStrs = append(finalStrs, line)
	}
	return finalStrs
}

func getShortBody(data string) []string {
	removedTop := strings.Split(data, `<div class="js-novel-text p-novel__text">`)
	summBody := strings.Split(removedTop[1], `</div>`)
	lines := strings.Split(summBody[0], `</p>`)
	reg := regexp.MustCompile(`<\s*(.*?)\s*>`)
	var finalStrs []string
	for _, line := range lines {
		line = reg.ReplaceAllString(line, "")
		if line == "" {
			continue
		}
		finalStrs = append(finalStrs, line)
	}
	return finalStrs
}

func GetChapterData(data string) (Chapter, error) {
	var lines []string
	titleFind := regexp.MustCompile(`<h1 class="p-novel__title p-novel__title--rensai">\s*(.*?)\s*</h1>`)
	title := titleFind.FindAllStringSubmatch(data, 1)
	topCrop := strings.Split(data, `<div class="js-novel-text p-novel__text">`)
	if len(topCrop) < 2 {
		panic("Couldnt find body, possible bad request")
	}
	botCrop := strings.Split(topCrop[1], "</div>")
	filteredBody := strings.Split(botCrop[0], "</p>")
	reg := regexp.MustCompile(`<\s*(.*?)\s*>`)
	for _, line := range filteredBody {
		res := reg.ReplaceAllString(line, "")
		clean := strings.TrimSpace(res)
		lines = append(lines, clean)
	}
	if len(title[0]) < 2 {
		return Chapter{}, errors.New("Regex failed to find title")
	}
	return Chapter{
		StatusCode: 200,
		Title:      title[0][1],
		Contents:   lines,
	}, nil

}

func (c *Client) GetFromRange(urlData UrlInfo, chapterAmount int) ([]Chapter, error) {
	//var userAgents = []string{
	//"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	//"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	//"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	//}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0"
	downloadBar := progressbar.Default(int64(chapterAmount), "Downloading chapters...\n")
	var scrappedChapters []Chapter
	for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
		chapter, err := c.GetSingleAddress(fullURL, userAgent)
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
