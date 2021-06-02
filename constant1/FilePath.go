package constant1

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//修改回来
var PARENT_DIR_PATH string = GetCurrentPath()

var ALGOLIA_COMPLIE_JSON_PATH string = PARENT_DIR_PATH + "/public/algolia.json"
var CACHE_ALGOLIA_JSON_PATH string = PARENT_DIR_PATH + "/cache_algolia.json"
var MD5_ALGOLIA_JSON_PATH string = PARENT_DIR_PATH + "/md5_algolia.json"
var Num int32 = 0

func init() {
	fmt.Println("current path:" + GetCurrentPath())
}

//func GetCurrentFilePath() string {
////	_, filePath, _, _ := runtime.Caller(1)
////	return filePath
////}

func GetCurrentPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}