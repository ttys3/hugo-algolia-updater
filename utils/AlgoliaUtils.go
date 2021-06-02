package utils

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
)

var HttpProxy = ""

// 更新分词
func UpdateAlgolia(objects []algoliasearch.Object) bool {
	appID := os.Getenv("ALG_APP_ID")
	appKey := os.Getenv("ALG_APP_KEY")
	indexName := os.Getenv("ALG_INDEX")
	client := algoliasearch.NewClient(appID, appKey)
	if HttpProxy != "" {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse("http://127.0.0.1:1087")
			// return url.Parse("ss://rc4-md5:123456@ss.server.com:1080")
		}
		tr := &http.Transport{Proxy: proxy}
		httpclient := &http.Client{
			Transport: tr,
		}
		client.SetHTTPClient(httpclient)
	}

	log.Printf("begin re-index %s\n", indexName)
	index := client.InitIndex(indexName)
	index.Clear()
	index.AddObjects(objects)
	return true
}
