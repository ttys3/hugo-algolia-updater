package common

import (
	"net/http"
	"net/url"
	"os"

	"go.uber.org/zap"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
)

var HttpProxy = ""

// 更新分词
func UpdateAlgolia(indexName, appID, adminKey string, objects []algoliasearch.Object) error {
	// allow override from env vars
	if id := os.Getenv("ALG_APP_ID"); id != "" {
		appID = id
		zap.S().Infof("override appID from env var, appID=%v", appID)
	}
	if key := os.Getenv("ALG_APP_KEY"); key != "" {
		adminKey = key
		zap.S().Infof("override adminKey from env var, adminKey=%v", adminKey)
	}

	if idx := os.Getenv("ALG_INDEX"); idx != "" {
		indexName = idx
		zap.S().Infof("override indexName from env var, indexName=%v", indexName)
	}

	client := algoliasearch.NewClient(appID, adminKey)
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

	zap.S().Infof("begin re-index, indexName=%v appID=%v", indexName, appID)

	index := client.InitIndex(indexName)
	if _, err := index.Clear(); err != nil {
		return err
	}
	if _, err := index.AddObjects(objects); err != nil {
		return err
	}
	return nil
}
