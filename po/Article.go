package po

type Article struct {
	Yaml        HugoJsonPost
	Content     string
	Md5Value    string
	Participles *[]string
}
