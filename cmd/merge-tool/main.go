package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/ttys3/hugo-algolia-updater/common"
)

var (
	pathsWithCommaSep string
	outfile           string
)

func main() {
	flag.StringVar(&pathsWithCommaSep, "f", "", "files with comma separated to merge")
	flag.StringVar(&outfile, "t", "out.dict", "destination file to save to")
	flag.Parse()

	if pathsWithCommaSep == "" {
		os.Exit(-1)
	}

	// example dict  "/github.com/go-ego/gse/data/dict/dictionary.txt,/github.com/go-ego/gse/data/dict/zh/dict.txt,
	// github.com/yanyiwu/gojieba/dict/jieba.dict.utf8"
	// example stop words "/github.com/yanyiwu/gojieba/dict/stop_words.utf8,/Hugo/Naah-Blog/stop.txt"
	mergeDict(pathsWithCommaSep, outfile)
}

func mergeDict(pathsWithCommaSep string, outfile string) {
	split := strings.Split(pathsWithCommaSep, ",")
	set := mapset.NewSet()
	for index1 := range split {
		p := split[index1]
		s := common.ReadFileString(p)
		sa := strings.Split(s, "\r\n")
		for index2 := range sa {
			set.Add(sa[index2])
			common.Num++
		}

	}
	slice := set.ToSlice()
	fmt.Print(len(slice))
	array := common.InterfaceArray2StringArray(slice, 1)
	str := strings.Join(array, "\n")
	common.WriteFile(outfile, []byte(str))
}
