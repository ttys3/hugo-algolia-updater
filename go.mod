module github.com/ttys3/hugo-algolia-updater

go 1.16

replace github.com/yanyiwu/gojieba v1.1.2 => github.com/ttys3/gojieba v1.1.3

require (
	github.com/algolia/algoliasearch-client-go v2.25.0+incompatible
	github.com/deckarep/golang-set v1.7.1
	github.com/go-ego/gse v0.67.0
	github.com/json-iterator/go v1.1.11
	github.com/yanyiwu/gojieba v1.1.2
	go.uber.org/zap v1.17.0
)
