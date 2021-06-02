package common

import (
	"fmt"
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
			return url.Parse(HttpProxy)
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
	zap.S().Infof("clear index, indexName=%v appID=%v", indexName, appID)
	if _, err := index.Clear(); err != nil {
		return err
	}
	_, err := index.SetSettings(algoliasearch.Map{
		"searchableAttributes": []string{
			"unordered(title)",
			"unordered(keywords)",
			"unordered(description)",
			"unordered(content)",
			"url", // there's no ordered
		},
	})
	if err != nil {
		return fmt.Errorf("index.SetSettings searchableAttributes failed, err=%w", err)
	}
	zap.S().Infof("begin add objects to index, indexName=%v appID=%v objects count=%v", indexName, appID, len(objects))
	if _, err := index.AddObjects(objects); err != nil {
		return err
	}
	zap.S().Infof("done add objects to index, indexName=%v appID=%v objects count=%v", indexName, appID, len(objects))
	return nil
}
