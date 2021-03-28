package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	ChapterURL        = "/home/arieldalessandro/dump-wiki-one-piece/chapter_%d"
	CoverQuery        = "#Cover_Page"
	ShortSummaryQuery = "#Short_Summary"
	LongSummaryQuery  = "#Long_Summary"
)

type LocalTag struct {
	Url   string
	Alias string
}

type Chapter struct {
	gorm.Model
	Number       uint8
	Title        string
	ReleaseDate  time.Time
	CoverPage    string `grom:"type:text"`
	CoverURL     string
	ShortSummary string `grom:"type:text"`
	LongSummary  string `grom:"type:text"`
	WikiURL      string
	MangaURL     string
	Tags         []Tag `gorm:"many2many:chapter_tags"`
}

type Tag struct {
	gorm.Model
	URL      string
	Category string
	Chapters []Chapter `gorm:"many2many:chapter_tags"`
}

type Alias struct {
	gorm.Model
	Name  string
	TagID uint
	Tag   Tag
}

type ChapterTags struct {
	ChapterID uint   `gorm:"primaryKey;autoIncrement:false"`
	TagID     uint   `gorm:"primaryKey;autoIncrement:false"`
	Section   string `gorm:"primaryKey"`
}

func TestScrape(db *gorm.DB, chapterNumber int) {
	fmt.Println("Procesando el chapter", chapterNumber)
	file, err := os.Open(fmt.Sprintf(ChapterURL, chapterNumber))
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	var chapterModel = new(Chapter)

	cover, coverTags := ParseSection(doc, CoverQuery)
	shortSummary, shortSummaryTags := ParseSection(doc, ShortSummaryQuery)
	longSummary, longSummaryTags := ParseSection(doc, LongSummaryQuery)

	chapterModel.Number = uint8(chapterNumber)

	chapterModel.CoverPage = cover
	chapterModel.ShortSummary = shortSummary
	chapterModel.LongSummary = longSummary

	var charTags []LocalTag
	doc.Find(".CharTable a").Each(func(i int, s *goquery.Selection) {
		var tag = new(LocalTag)
		tag.Url, _ = s.Attr("href")
		tag.Alias = s.Text()
		charTags = append(charTags, *tag)
	})

	generalInfo := doc.Find("#mw-content-text > div > aside")

	title := generalInfo.Find("[data-source=title]").Text()

	chapterModel.Title = title

	releaseDate := generalInfo.Find("[data-source=date2] > div").Text()
	releaseDate = strings.Trim(releaseDate, "[ref]")
	relDate, _ := time.Parse("January 2, 2006", releaseDate)

	chapterModel.ReleaseDate = relDate

	coverUrl, _ := generalInfo.Find("[data-source=image] > a").Attr("href")

	chapterModel.CoverURL = coverUrl

	db.Create(&chapterModel)

	CreateTags(chapterModel, coverTags, "cover", db)
	CreateTags(chapterModel, shortSummaryTags, "short", db)
	CreateTags(chapterModel, longSummaryTags, "long", db)
	CreateTags(chapterModel, charTags, "characters", db)

	fmt.Println("Terminand el chapter", chapterNumber)
}

func CreateTags(chapter *Chapter, tags []LocalTag, section string, db *gorm.DB) {
	for _, tag := range tags {
		var tagModel = new(Tag)
		tagModel.URL = tag.Url
		db.FirstOrCreate(&tagModel, Tag{URL: tag.Url})

		var aliasModel = new(Alias)
		aliasModel.Name = tag.Alias
		aliasModel.TagID = tagModel.ID
		db.FirstOrCreate(&aliasModel, Alias{Name: tag.Alias, TagID: tagModel.ID})

		var chapterTag = new(ChapterTags)
		chapterTag.ChapterID = chapter.ID
		chapterTag.TagID = tagModel.ID
		chapterTag.Section = section

		db.FirstOrCreate(&chapterTag, ChapterTags{ChapterID: chapter.ID, TagID: tagModel.ID, Section: section})
	}
}

func ParseSection(doc *goquery.Document, query string) (string, []LocalTag) {
	node := doc.Find(query).Parent().Next()

	var text string
	var tags []LocalTag

	for node.Is("p") {
		text += node.Text()
		node.Find("a").Each(func(i int, s *goquery.Selection) {
			var tag = new(LocalTag)
			tag.Url, _ = s.Attr("href")
			tag.Alias = s.Text()
			tags = append(tags, *tag)
		})

		node = node.Next()
	}

	return text, tags
}

func main() {
	db, err := gorm.Open(mysql.Open("root:secret@tcp(localhost:3306)/op_api?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.SetupJoinTable(&Chapter{}, "Tags", &ChapterTags{})
	if err != nil {
		panic("Fail to set up join")
	}

	err = db.SetupJoinTable(&Tag{}, "Chapters", &ChapterTags{})
	if err != nil {
		panic("Fail to set up join")
	}

	db.AutoMigrate(&Chapter{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Alias{})
	db.AutoMigrate(&ChapterTags{})

	for i := 1; i < 1008; i++ {
		TestScrape(db, i)
	}
}
