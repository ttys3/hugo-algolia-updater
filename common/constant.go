package common

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 修改回来
var PARENT_DIR_PATH string = GetCurrentPath()

var (
	ALGOLIA_COMPLIE_JSON_PATH string = PARENT_DIR_PATH + "/public/algolia.json"
	CACHE_ALGOLIA_JSON_PATH   string = PARENT_DIR_PATH + "/cache_algolia.json"
	MD5_ALGOLIA_JSON_PATH     string = PARENT_DIR_PATH + "/md5_algolia.json"
	Num                       int32  = 0
)

func GetCurrentPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
