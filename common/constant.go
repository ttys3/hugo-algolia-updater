package common

import (
	"log"
	"os"
	"strings"
)

var WORKING_DIR_PATH = GetWorkingDir()

var (
	ALGOLIA_COMPLIE_JSON_PATH       = WORKING_DIR_PATH + "/public/algolia.json"
	CACHE_ALGOLIA_JSON_PATH         = WORKING_DIR_PATH + "/algolia.cache.json"
	MD5_ALGOLIA_JSON_PATH           = WORKING_DIR_PATH + "/algolia.md5.json"
	HUGO_INDEX_JSON_PATH            = "public/index.json"
	Num                       int32 = 0
)

func GetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
