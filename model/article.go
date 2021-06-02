package model

type Article struct {
	HugoJsonPost HugoJsonPost
	Content      string
	Md5Value     string
	Participles  *[]string
}
