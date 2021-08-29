package main

import "time"

type Chapter struct {
	Number       uint        `json:"number"`
	Title        string      `json:"title"`
	ReleaseDate  time.Time   `json:"release_date"`
	Links        []Link      `json:"links"`
	Cover        Cover       `json:"cover"`
	ShortSummary Summary     `json:"short_summary"`
	Summary      Summary     `json:"summary"`
	Characters   []Reference `json:"characters"`
}

type Link struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cover struct {
	Text       string      `json:"text"`
	Image      string      `json:"image"`
	References []Reference `json:"references"`
}

type Summary struct {
	Text       string      `json:"text"`
	References []Reference `json:"references"`
}

type Reference struct {
	Name string `json:"name"`
	Wiki string `json:"wiki"`
}
