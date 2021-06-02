package model

import "github.com/algolia/algoliasearch-client-go/algoliasearch"

type Algolia struct {
	ObjectID    string   `json:"objectID"`
	Title       string   `json:"title"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	// url seems the same as pathname
	URL string `json:"url"`
	// Pathname    string `json:"pathname"`
	// UrlDepth    int		 `json:"urlDepth"`
	// Position int `json:"position"`
	Lang          string `json:"lang"`
	Origin        string `json:"origin"`
	Image         string `json:"image"`
	DatePublished int64  `json:"datePublished"`

	Subtitle   string   `json:"subtitle"`
	Date       string   `json:"date"`
	Author     string   `json:"author"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
}

func (a *Algolia) ToMap() algoliasearch.Object {
	m := algoliasearch.Object{}
	m["objectID"] = a.ObjectID
	m["title"] = a.Title
	m["keywords"] = a.Keywords
	m["description"] = a.Description
	m["content"] = a.Content
	m["url"] = a.URL
	m["lang"] = a.Lang
	m["origin"] = a.Origin
	m["image"] = a.Image
	m["datePublished"] = a.DatePublished
	m["subtitle"] = a.Subtitle
	m["date"] = a.Date
	m["author"] = a.Author
	m["tags"] = a.Tags
	m["categories"] = a.Categories
	return m
}

// demo algolia Standard record schema from https://www.algolia.com/doc/tools/crawler/netlify-plugin/extraction-strategy/#default-schema
/*
objectID  "https://example.com/post/frontend/basis/why-firefox-option-background-color-do-not-work-under-linux/#0"
title "为什么 Firefox option 的 background-color 样式在 Linux 下不工作 :: /dev/ttyS3"
keywords  [ "荒野無燈", "golang" ]
description  "系统版本： Fedora 32 Firefox版本： 76.0.1 (64-bit) 我偶然发现一个页面bug。Chromium下是显示正常的： 而Firefox下是这样的： 这个HTML代码"
content  "为什么 Firefox option 的 background-color 样式在 Linux 下不工作 2020-05-20 :: 荒 …"
url  "/post/frontend/basis/why-firefox-option-background-color-do-not-work-under-linux/"
urlDepth 4
position 0
lang "en"
origin "https://example.com"
pathname "/post/frontend/basis/why-firefox-option-background-color-do-not-work-under-linux/"
image "https://example.com/img/favicon/favicon-96x96.png"
datePublished 1590040593
*/
