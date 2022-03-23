package main

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"io/ioutil"
	"strings"
)

// Colly is a Golang framework for building web scrapers. Docs is here: http://go-colly.org/docs/.
// I'll use colly to build a scraper to fetch infos from https://chd.web.sdo.com/web7/news/newslist.asp?CategoryID=306.
const domain = "https://chd.web.sdo.com/web7"
const latestNewsLinkFileName = "./latest_news_link"

func Fetch() (link, content string) {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnHTML(".news-list ul li:nth-of-type(1) p a[href]", func(a *colly.HTMLElement) {
		link = strings.ReplaceAll(a.Attr("href"), "..", domain)
		content = strings.TrimSpace(a.Text)
	})

	_ = c.Visit(domain + "/news/newslist.asp?CategoryID=306")
	return
}

func readLink() string {
	content, err := ioutil.ReadFile(latestNewsLinkFileName)
	if err != nil {
		return ""
	}
	return string(content)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeLink(id string) {
	content := []byte(id)
	err := ioutil.WriteFile(latestNewsLinkFileName, content, 0644)
	if err != nil {
		panic(err)
	}
}
