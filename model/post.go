package model

import "time"

type HugoJsonPost struct {
	Title      string      `json:"title"`
	Images []string `json:"images"`
	Tags       []string    `json:"tags"`
	Categories []string `json:"categories"`
	Content   string      `json:"content"`
	Description string `json:"description"`
	Permalink  string      `json:"permalink"`
	Date       time.Time   `json:"date"`
	Section    string      `json:"section"`
}
