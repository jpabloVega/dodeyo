package cmd

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"uuid"

	"github.com/jpabloVega/dodeyo/api"
	"github.com/jpabloVega/dodeyo/database"
	"github.com/jpabloVega/dodeyo/email"
	"github.com/raitucarp/epub"
	"github.com/schollz/progressbar/v3"
	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LINEHEIGHT    = 20.0
	MAXPAGEHEIGHT = 750.0
)

var (
	bodySize        int
	titleSize       int
	destinationPath string
	fileName        string
	oPDF            bool
	oHTML           bool
	oEPub           bool
	vertical        bool
	mailToKindle    bool
	replaceKanji    bool
	chapterAmount   int
	startFrom       int
	AssetsFS        embed.FS
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVar(&bodySize, "bodysize", 14, "Size for the font of the body (only for pdf)")
	getCmd.Flags().IntVar(&titleSize, "titlesize", 16, "Size for the font of the titles (only for pdf)")
	getCmd.Flags().IntVarP(&chapterAmount, "amount", "a", 10, "Download this amount of chapters")
	getCmd.Flags().IntVar(&startFrom, "start", 0, "Start from chapter")
	getCmd.Flags().StringVarP(&destinationPath, "saveto", "s", "", "Path to the folder where pdf's are stored")
	getCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the newly created pdf, defaults to novel title")
	getCmd.Flags().BoolVarP(&vertical, "vertical", "v", false, "Export text in vertical(only availible to html and epub)")
	getCmd.Flags().BoolVar(&mailToKindle, "send", false, "Mail to kindle")
	getCmd.Flags().BoolVar(&replaceKanji, "furigana", false, "Add furigana")
	getCmd.Flags().BoolVar(&oPDF, "pdf", false, "Output in PDF")
	getCmd.Flags().BoolVar(&oHTML, "html", false, "Output in HTML")
	getCmd.Flags().BoolVar(&oEPub, "epub", false, "Output in EPUB")
	getCmd.MarkFlagsOneRequired("pdf", "html", "epub")
	getCmd.MarkFlagsMutuallyExclusive("pdf", "html", "epub")

}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Export novel from syosetsu",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE:  GetContents,
}

func GetContents(cmd *cobra.Command, args []string) error {
	c := api.NewClient(5 * time.Second)

	for _, address := range args {
		err := GetCommand(cmd, address, &c)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCommand(cmd *cobra.Command, address string, c *api.Client) error {
	var err error
	if destinationPath == "" {
		destinationPath = viper.GetString("defaultpath")
		if destinationPath == "" {
			destinationPath, err = setDestinationPath()
			if err != nil {
				return err
			}
		}
	}

	var urlData api.UrlInfo
	urlData, err = api.GetSiteInfo(address)
	if err != nil {
		return err
	}

	db, err := database.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	bookExists, dbBook, err := database.GetBook(db, urlData.Url)
	if bookExists {
		urlData.StartIndex = dbBook.LastDownload
	}

	if startFrom != 0 {
		urlData.StartIndex = startFrom
	}

	var bookCredits api.Credits
	bookCredits, err = c.GetChapterAuthor(urlData.Url)
	if err != nil {
		return err
	}

	var chapters []api.Chapter

	switch bookCredits.Type {
	case api.Short:
		var statusCode int
		if bookCredits.Title != "" && len(bookCredits.Body) != 0 {
			statusCode = 200
		}
		onlyChap := api.Chapter{
			StatusCode: statusCode,
			Title:      bookCredits.Title,
			Contents:   bookCredits.Body,
		}
		chapters = append(chapters, onlyChap)
	case api.Serialized:
		chapters, err = c.GetFromRange(urlData, chapterAmount)
		if err != nil {
			return err
		}
	default:
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
		return errors.New("No chapters to create document")
	}

	if replaceKanji {
		furiChap := []api.Chapter{}
		for _, chap := range chaptersData {
			repChap, err := api.ReplaceKanji(chap)
			if err != nil {
				return err
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
		return err
	}

	lastDownload := urlData.StartIndex + len(chaptersData)

	newBook := database.DBBook{
		Name:         bookCredits.Title,
		Link:         urlData.Url,
		LastDownload: lastDownload,
		NovelType:    bookCredits.Type,
		DownloadedAt: time.Now(),
	}
	err = database.SaveBook(db, newBook)
	if err != nil {
		return err
	}

	docs := []string{path}
	if mailToKindle {
		err = email.MailToKindle(docs)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetHTML(chaptersData []api.Chapter, bookCredits api.Credits) (string, error) {
	htmlProgress := progressbar.Default(int64(len(chaptersData)), "Creating html...\n")

	var vertFormat string
	if vertical {
		vertFormat = `
			writing-mode: vertical-rl;
            -webkit-writing-mode: vertical-rl;`
	}

	var html strings.Builder
	header := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>` + bookCredits.Title + `</title>
    <style>
        body {
            font-family: serif;
            padding: 0;
            margin: 0;` + vertFormat + `
        }
        p {
            text-indent: 0;
            margin-top: 0;
            margin-bottom: 0;
            text-align: justify;
        }
        h1, h2, h3 {
            text-align: center;
            text-indent: 0;
            margin-top: 1em;
            margin-bottom: 0.5em;
        }
    </style>
</head>
<body>

`

	html.WriteString(header)
	if bookCredits.Type == api.Serialized && len(bookCredits.Body) != 0 {
		for _, line := range bookCredits.Body {
			if line == "" {
				continue
			}
			fmt.Fprintf(&html, `    <p>%s</p>
`, line)
		}
		html.WriteString(`

    <mbp:pagebreak/>

`)
	}
	for _, chapter := range chaptersData {
		htmlProgress.Add(1)
		fmt.Fprintf(&html, `	<h1>%s</h1>

`, chapter.Title)
		for _, body := range chapter.Contents {
			if body == "" {
				continue
			}
			fmt.Fprintf(&html, `    <p>%s</p>
`, body)
		}
		html.WriteString(`

    <mbp:pagebreak/>

`)
		htmlProgress.Add(1)
	}

	chapterTitle := fmt.Sprintf("書名： %s", bookCredits.Title)
	authorCredit := fmt.Sprintf("作者： %s", bookCredits.Author)
	chapterUrl := fmt.Sprintf("URL： %s", bookCredits.URL)

	fmt.Fprintf(&html, `    <h1>%s</h1>
	<h1>%s</h1>
    <h1>%s</h1>
</body>
</html>`, chapterTitle, authorCredit, chapterUrl)

	filePath := getOrigName(destinationPath, fileName, bookCredits.Title, "html")

	err := os.WriteFile(filePath, []byte(html.String()), 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("Successfully created HTML at %s\n", filePath)
	return filePath, nil
}

func GetPDF(chapterData []api.Chapter, credits api.Credits) (string, error) {
	pdfProgress := progressbar.Default(int64(len(chapterData)), "Creating PDF...\n")

	// Create pdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	regularFont, err := AssetsFS.Open("assets/NotoSansJP-Regular.ttf")
	if err != nil {
		return "", err
	}
	defer regularFont.Close()

	boldFont, err := AssetsFS.Open("assets/NotoSansJP-Bold.ttf")
	if err != nil {
		return "", err
	}
	defer boldFont.Close()

	err = pdf.AddTTFFontByReaderWithOption("NotoSansJP", regularFont, gopdf.TtfOption{Style: gopdf.Regular})
	if err != nil {
		return "", err
	}
	err = pdf.AddTTFFontByReaderWithOption("NotoSansJP", boldFont, gopdf.TtfOption{Style: gopdf.Bold})
	if err != nil {
		return "", err
	}

	if credits.Type == api.Serialized && len(credits.Body) != 0 {
		pdf.AddPage()
		pdf.SetX(50)

		err = pdf.SetFont("NotoSansJP", "B", titleSize)
		if err != nil {
			return "", err
		}

		bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
		err = pdf.MultiCellWithOption(bodyRect, "あらすじ", gopdf.CellOption{Align: gopdf.Center})
		if err != nil {
			return "", err
		}
		pdf.Br(40)
		err = pdf.SetFont("NotoSansJP", "", bodySize)
		if err != nil {
			return "", err
		}
		for _, line := range credits.Body {
			if line == "" {
				continue
			}
			if pdf.GetY() > MAXPAGEHEIGHT {
				pdf.AddPage()
				pdf.SetX(50)
				_ = pdf.SetFont("NotoSansJP", "", bodySize)
			}
			pdf.SetX(50)
			bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
			err = pdf.MultiCellWithOption(bodyRect, line, gopdf.CellOption{Align: gopdf.Left})
			if err != nil {
				return "", err
			}
		}
	}

	for _, chapter := range chapterData {
		pdfProgress.Add(1)
		pdf.AddPage()
		pdf.SetX(50)

		err = pdf.SetFont("NotoSansJP", "B", titleSize)
		if err != nil {
			return "", err
		}

		// Set title
		bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
		err = pdf.MultiCellWithOption(bodyRect, chapter.Title, gopdf.CellOption{Align: gopdf.Center})
		if err != nil {
			return "", err
		}
		pdf.Br(40)

		// Set body
		err = pdf.SetFont("NotoSansJP", "", bodySize)
		if err != nil {
			return "", err
		}

		for _, line := range chapter.Contents {
			if line == "" {
				continue
			}

			if pdf.GetY() > MAXPAGEHEIGHT {
				pdf.AddPage()
				pdf.SetX(50)
				_ = pdf.SetFont("NotoSansJP", "", bodySize)
			}
			pdf.SetX(50)
			bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
			err = pdf.MultiCellWithOption(bodyRect, line, gopdf.CellOption{Align: gopdf.Left})
			if err != nil {
				return "", err
			}
		}
	}
	pdf.AddPage()
	pdf.SetX(50)
	err = pdf.SetFont("NotoSansJP", "B", titleSize)
	if err != nil {
		return "", err
	}

	chapterTitle := fmt.Sprintf("書名： %s", credits.Title)
	bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}

	err = pdf.MultiCellWithOption(bodyRect, chapterTitle, gopdf.CellOption{Align: gopdf.Center})
	if err != nil {
		return "", err
	}
	pdf.Br(40)

	pdf.SetX(50)
	authorCredit := fmt.Sprintf("作者： %s", credits.Author)
	err = pdf.MultiCellWithOption(bodyRect, authorCredit, gopdf.CellOption{Align: gopdf.Center})
	if err != nil {
		return "", err
	}
	pdf.Br(40)

	pdf.SetX(50)
	chapterUrl := fmt.Sprintf("URL： %s", credits.URL)
	err = pdf.MultiCellWithOption(bodyRect, chapterUrl, gopdf.CellOption{Align: gopdf.Center})
	if err != nil {
		return "", err
	}
	pdf.Br(40)

	filePath := getOrigName(destinationPath, fileName, credits.Title, "pdf")

	err = pdf.WritePdf(filePath)
	if err != nil {
		return "", err
	}

	fmt.Printf("Pdf successfully created at: %s\n", filePath)
	return filePath, nil
}

func GetEPub(chapterData []api.Chapter, credits api.Credits) (string, error) {
	ePubProgress := progressbar.Default(int64(len(chapterData)), "Creating Epub...\n")
	id := uuid.New()

	w := epub.New(fmt.Sprintf("urn:uuid:%s", id.String()))

	w.Title(credits.Title)
	w.Author(credits.Author)
	w.Languages("ja")

	var cssStyling string
	var vertHeader string

	if vertical {
		w.Direction("rtl")
		vertHeader = ` dir="rtl"`
		cssStyling = `
	<style>
		html {
			writing-mode: vertical-rl;
			-webkit-writing-mode: vertical-rl;
			-epub-writing-mode: vertical-rl;
			font-family: "serif", "Yu Mincho", "MS Mincho", serif;
		}
		p {
			text-indent: 1em;
			margin: 0;
			line-height: 1.8;
		}
		.tate-chu-yoko {
			-webkit-text-combine: horizontal;
			-epub-text-combine: horizontal;
			text-combine-upright: all;
		}
	</style>`
	} else {
		cssStyling = `
	<style>
		html {
			font-family: "serif", "Yu Mincho", "MS Mincho", serif;
		}
		p {
			text-indent: 1em;
			margin: 0;
			line-height: 1.8;
		}
	</style>`
	}

	w.MetaContent(map[string]string{
		"rendition:layout":      "reflowable",
		"rendition:orientation": "auto",
	})

	var tocItems []epub.TOC

	var xhtml strings.Builder

	if credits.Type == api.Serialized && len(credits.Body) != 0 {
		curr := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://w3.org"%s xml:lang="ja" lang="ja">
<head>
    <title>%s</title>
%s
</head>
<body>
	<h1>%s</h1>
`, vertHeader, credits.Title, cssStyling, credits.Title)
		xhtml.WriteString(curr)
		for _, line := range credits.Body {
			if line == "" {
				continue
			}
			currLine := fmt.Sprintf(`	<p>%s</p>
`, line)
			xhtml.WriteString(currLine)

		}
		endLines := `</body>
</html>`
		xhtml.WriteString(endLines)

		w.AddContent("あらすじ", []byte(xhtml.String()))

		tocItems = append(tocItems, epub.TOC{
			Title: credits.Title,
			Href:  "あらすじ",
		})
	}

	for i, chapter := range chapterData {
		ePubProgress.Add(1)
		currChapter := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://w3.org"%s xml:lang="ja" lang="ja">
<head>
    <title>%s</title>
%s
</head>
<body>
	<h1>%s</h1>
`, vertHeader, chapter.Title, cssStyling, chapter.Title)
		xhtml.WriteString(currChapter)
		for _, line := range chapter.Contents {
			if line == "" {
				continue
			}
			currLine := fmt.Sprintf(`	<p>%s</p>
`, line)
			xhtml.WriteString(currLine)
		}

		endLines := `</body>
</html>`
		xhtml.WriteString(endLines)
		chapterName := fmt.Sprintf("Chapter-%d.xhtml", i)

		w.AddContent(chapterName, []byte(xhtml.String()))

		tocItems = append(tocItems, epub.TOC{
			Title: chapter.Title,
			Href:  chapterName,
		})
	}

	mainTOC := epub.TOC{
		Title: "目次",
		Items: tocItems,
	}

	err := w.TableOfContents("toc", mainTOC)
	if err != nil {
		return "", err
	}

	filePath := getOrigName(destinationPath, fileName, credits.Title, "epub")

	err = w.Write(filePath)
	if err != nil {
		return "", err
	}

	fmt.Printf("Successfully created epub at: %s\n", filePath)
	return filePath, nil
}

func getOrigName(destPath, fileName, title, ext string) string {
	if fileName == "" {
		fileName = title
	}
	if title == "" {
		title = "untitled"
	}
	counter := 2
	fileNameExt := fmt.Sprintf("%s.%s", fileName, ext)
	filePath := filepath.Join(destPath, fileNameExt)
	for {
		_, err := os.Stat(filePath)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		fileNameExt = fmt.Sprintf("%s-%d.%s", fileName, counter, ext)
		filePath = filepath.Join(destPath, fileNameExt)
		counter++
	}
	return filePath
}

func setDestinationPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter destination path > ")

	inputPath, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(inputPath)

	err = checkValidPath(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func checkValidPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return fmt.Errorf("Error creating directory:%v\n", err)
			}
		}
		return fmt.Errorf("Error accessing path: %v\n", err)
	}

	if !info.IsDir() {
		return errors.New("Path doesnt point to a directory")
	}

	viper.Set("defaultpath", path)
	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
