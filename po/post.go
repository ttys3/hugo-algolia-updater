package po

import "time"

type HugoJsonPost struct {
	Categories interface{} `json:"categories"`
	Contents   string      `json:"contents"`
	Date       time.Time   `json:"date"`
	Permalink  string      `json:"permalink"`
	Section    string      `json:"section"`
	Tags       []string    `json:"tags"`
	Title      string      `json:"title"`
}