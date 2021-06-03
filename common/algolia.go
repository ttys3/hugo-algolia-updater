package common

import (
	"fmt"
	"os"
	"time"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/ttys3/hugo-algolia-updater/model"
	"go.uber.org/zap"

	algoliasearch "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

// UpdateAlgolia 清除索引并重新添加索引数据
func UpdateAlgolia(indexName, appID, adminKey string, objects []*model.Algolia) error {
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

	// https://www.algolia.com/doc/api-client/getting-started/upgrade-guides/go/#upgrade-from-v2-to-v3
	// algolia proxy using go standard lib env var HTTP_PROXY or HTTPS_PROXY via http.ProxyFromEnvironment
	client := algoliasearch.NewClientWithConfig(algoliasearch.Configuration{
		AppID:        appID,            // Mandatory
		APIKey:       adminKey,         // Mandatory
		ReadTimeout:  5 * time.Second,  // Optional
		WriteTimeout: 10 * time.Second, // Optional
	})

	zap.S().Infof("begin re-index, indexName=%v appID=%v", indexName, appID)
	index := client.InitIndex(indexName)
	zap.S().Infof("clear index, indexName=%v appID=%v", indexName, appID)
	if _, err := index.ClearObjects(); err != nil {
		return err
	}

	// see https://www.algolia.com/doc/api-reference/api-parameters/searchableAttributes/
	_, err := index.SetSettings(algoliasearch.Settings{
		SearchableAttributes: opt.SearchableAttributes(
			"unordered(title)",
			"unordered(keywords)",
			"unordered(description)",
			"unordered(content)",
			"url", // there's no ordered
		),
	})
	if err != nil {
		return fmt.Errorf("index.SetSettings searchableAttributes failed, err=%w", err)
	}
	zap.S().Infof("begin add objects to index, indexName=%v appID=%v objects count=%v", indexName, appID, len(objects))
	if _, err := index.SaveObjects(objects); err != nil {
		return err
	}
	zap.S().Infof("done add objects to index, indexName=%v appID=%v objects count=%v", indexName, appID, len(objects))
	return nil
}
