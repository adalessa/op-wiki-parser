package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	WikiUrl string = "https://onepiece.fandom.com/wiki/Chapter_%d"

	CoverQuery        string = "#Cover_Page"
	ShortSummaryQuery        = "#Short_Summary"
	LongSummaryQuery         = "#Long_Summary"

	ChapterURL string = "/home/arieldalessandro/dump-wiki-one-piece/chapter_%d"
	MangaUrl          = "https://manganelo.com/chapter/tkqu521609849722/chapter_%d"

	APIUrl string = "https://op-api.ad-impeldown.synology.me/api/chapters"
	// APIUrl string = "http://op-api.test/api/chapters"
)

func main() {
	var number uint = 1
	for number <= 1022 {
		chapter, err := processChapter(number)
		if err != nil {
			fmt.Println(number, err)
			return
		}
		err = pushToApi(*chapter)
		if err != nil {
			fmt.Println(number, err)
			return
		}
		number++
	}

}

func pushToApi(chapter Chapter) error {
	data, err := json.Marshal(chapter)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, APIUrl, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fmt.Println(resp.Status)
	io.Copy(os.Stdout, resp.Body)

	return nil
}

func processChapter(number uint) (*Chapter, error) {
	resp, err := HttpDownload(fmt.Sprintf(WikiUrl, number))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	chapter := Chapter{}

	chapter.Number = number

	generalInfo := doc.Find("#mw-content-text > div > aside")
	chapter.Title = generalInfo.Find("[data-source=title]").Text()

	releaseDate := generalInfo.Find("[data-source=date2] > div").Text()
	releaseDate = strings.Trim(releaseDate, "[ref]")
	relDate, _ := time.Parse("January 2, 2006", releaseDate)
	chapter.ReleaseDate = relDate

	chapter.Links = []Link{
		{
			Name:  "wiki",
			Value: fmt.Sprintf(WikiUrl, number),
		},
		{
			Name:  "manganelo",
			Value: fmt.Sprintf(MangaUrl, number),
		},
	}

	cover, coverRefs := parseSection(doc, CoverQuery)
	coverUrl, exists := generalInfo.Find("figure > a").Attr("href")
	if !exists {
		return nil, errors.New("Cover image does not exists")
	}
	if cover == "" {
		cover = "Not Available"
	}

	chapter.Cover.Text = cover
	chapter.Cover.References = coverRefs
	chapter.Cover.Image = coverUrl

	shortSummary, shortSummaryRefs := parseSection(doc, ShortSummaryQuery)
	chapter.ShortSummary.Text = shortSummary
	chapter.ShortSummary.References = shortSummaryRefs

	summary, summaryRefs := parseSection(doc, LongSummaryQuery)
	chapter.Summary.Text = summary
	chapter.Summary.References = summaryRefs

	var refs []Reference
	doc.Find(".CharTable a").Each(func(i int, s *goquery.Selection) {
		var ref = new(Reference)
		ref.Wiki, _ = s.Attr("href")
		ref.Name = s.Text()
		refs = append(refs, *ref)
	})

	chapter.Characters = refs

	return &chapter, nil
}

func HttpDownload(link string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:78.0) Gecko/20100101 Firefox/78.0")
	client := &http.Client{}

	return client.Do(req)
}

type LocalTag struct {
	Url   string
	Alias string
}

func parseSection(doc *goquery.Document, query string) (text string, refs []Reference) {
	node := doc.Find(query).Parent().Next()

	for node.Is("p") {
		text += node.Text()
		node.Find("a").Each(func(i int, s *goquery.Selection) {
			var ref = new(Reference)
			ref.Wiki, _ = s.Attr("href")
			ref.Name = s.Text()
			if ref.Wiki == "/wiki/Belly" {
				ref.Name = "Belly"
			}
			if ref.Wiki == "/wiki/Belly#Other_Currencies" {
				ref.Name = "Belly Other currencies"
			}
			refs = append(refs, *ref)
		})

		node = node.Next()
	}

	text = strings.ReplaceAll(text, "\n", "")
	text = html.EscapeString(text)

	if text == "" {
		text = "Not Available"
	}

	return text, refs
}
