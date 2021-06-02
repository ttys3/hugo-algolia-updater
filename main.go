package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/json-iterator/go"
	"github.com/ttys3/hugo-algolia-updater/common"
	"github.com/ttys3/hugo-algolia-updater/model"
)

var (
	serviceName string
	version     string
	buildTime   string
)

var showVersion bool

func main() {
	fmt.Printf("%s %s %s @%s\n", serviceName, version, buildTime, runtime.Version())

	flag.BoolVar(&showVersion, "v", false, "show version and exit")
	flag.Parse()

	if showVersion {
		return
	}

	common.InitJieba()

	startTime := time.Now().UnixNano() / 1e6

	// 运行编译
	execHugoBuild()

	participlesStartTime := time.Now().UnixNano() / 1e6

	articleList := getArticleList()

	// 获取分词列表
	cacheAlgoliasList := getCacheAlgoliasList()
	taskNum := 0
	flag := true
	// 有缓存时
	if len(cacheAlgoliasList) != 0 {
		exists, _ := common.Exists(common.MD5_ALGOLIA_JSON_PATH)
		if exists {
			flag = false
			// 有md5map
			common.Md5Map = common.NewConcurrentMap(getMd5Map())

			for _, article := range articleList {
				sss := article
				permalink := sss.HugoJsonPost.Permalink
				value := common.Md5Map.GetValue(permalink)
				oldMd5 := ""
				if value != nil {
					oldMd5 = value.(string)
				}
				compare := strings.Compare(oldMd5, sss.Md5Value)
				if compare != 0 {
					common.Queue.Push(sss)
					common.NeedArticleList = append(common.NeedArticleList, sss)
					taskNum++
				}
			}
		}
	}

	// 没缓存时
	if flag {
		for _, article := range articleList {
			common.Queue.Push(article)
			common.NeedArticleList = append(common.NeedArticleList, article)
			taskNum++
		}
	}

	// 创建WaitGroup（java中的countdown）
	common.WaitGroup.Add(taskNum)

	// 创建线程池
	pool := new(common.ThreadPool)
	pool.Init(runtime.NumCPU(), taskNum)

	// 循环添加任务
	for i := 0; i < taskNum; i++ {
		pool.AddTask(ParticiplesAsynchronous)
	}
	pool.Start()

	// 主线程阻塞
	common.WaitGroup.Wait()
	pool.Stop()
	fmt.Println("participles success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-participlesStartTime, 10) + " ms")

	// 创建分词
	algoliaStartTime := time.Now().UnixNano() / 1e6
	for _, article := range common.NeedArticleList {
		common.CacheAlgoliasMap[article.HugoJsonPost.Permalink] = model.Algolia{Title: article.HugoJsonPost.Title}
		// cacheAlgoliasList = append(cacheAlgoliasList, model.Algolia{Title: value.HugoJsonPost.Title})
	}

	objArray := []algoliasearch.Object{}
	for permalink, algolias := range common.CacheAlgoliasMap {

		value := common.ArticleMap.GetValue(permalink)
		var article *model.Article
		if value != nil {
			article = value.(*model.Article)
		} else {
			log.Printf("ArticleMap.GetValue failed, permalink=%v\n", permalink)
			continue
		}
		common.Md5Map.AddData(permalink, article.Md5Value)

		mapObj := common.Struct2Map(article.HugoJsonPost)
		// fmt.Printf("Struct2Map %#v\n", mapObj)

		if article.Participles != nil {
			participlesArray := *article.Participles
			var buffer bytes.Buffer
			for _, str := range participlesArray {
				if common.NumberReg.Match([]byte(str)) {
					continue
				}
				buffer.WriteString(str)
				buffer.WriteString(" ")
			}
			join := buffer.String()
			mapObj["content"] = join
		} else {
			mapObj["content"] = algolias.Content
		}
		mapObj["objectID"] = article.HugoJsonPost.Permalink
		mapObj["uri"] = article.HugoJsonPost.Permalink

		objArray = append(objArray, mapObj)
	}
	fmt.Println("generate algolia index success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-algoliaStartTime, 10) + " ms")
	fmt.Println("generate algolia index num: ", common.Num)

	return
	uploadStartTime := time.Now().UnixNano() / 1e6
	// 更新分词
	common.UpdateAlgolia(objArray)
	fmt.Println("update algolia success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-uploadStartTime, 10) + " ms")
	saveStartTime := time.Now().UnixNano() / 1e6
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	algoliaBytes, _ := json.Marshal(objArray)
	md5Bytes, _ := json.Marshal(common.Md5Map.GetData())
	common.WriteFile(common.ALGOLIA_COMPLIE_JSON_PATH, algoliaBytes)
	common.WriteFile(common.CACHE_ALGOLIA_JSON_PATH, algoliaBytes)
	common.WriteFile(common.MD5_ALGOLIA_JSON_PATH, md5Bytes)

	fmt.Println("save cache success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-saveStartTime, 10) + " ms")
	fmt.Println("total : " + strconv.FormatInt((time.Now().UnixNano()/1e6)-startTime, 10) + " ms")
}

func getArticleList() []*model.Article {
	hugoJsonFile := "public/index.json"
	c, err := ioutil.ReadFile(hugoJsonFile)
	if err != nil {
		panic(err)
	}
	var posts []*model.HugoJsonPost
	if err := json.Unmarshal(c, &posts); err != nil {
		panic(err)
	}

	var articleList []*model.Article
	taskNum := len(posts)
	// 创建WaitGroup（java中的countdown）
	common.WaitGroup.Add(taskNum)

	// 设置cpu并行数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建线程池
	pool := new(common.ThreadPool)
	pool.Init(runtime.NumCPU(), taskNum)

	for _, post := range posts {
		// log.Printf("post=%#v", post)
		post1 := post
		pool.AddTask(func() error {
			article := model.Article{HugoJsonPost: *post1, Content: post1.Contents, Md5Value: common.Md5V(post1.Contents)}
			articleList = append(articleList, &article)
			common.ArticleMap.AddData(post1.Permalink, &article)
			common.WaitGroup.Done()
			return nil
		})
	}

	pool.Start()
	// 主线程阻塞
	common.WaitGroup.Wait()
	pool.Stop()
	return articleList
}

// 多线程分词
func ParticiplesAsynchronous() error {
	article := common.Queue.Pop().(*model.Article)
	content := article.Content
	mdConf := article.HugoJsonPost

	participles := common.Participles(mdConf.Title, content)
	article.Participles = &participles
	fmt.Println("generate success: " + article.HugoJsonPost.Permalink)
	common.WaitGroup.Done()
	return nil
}

// 执行编译
func execHugoBuild() {
	out, _ := common.ExecShell("hugo", "--gc", "--minify", "--enableGitInfo")
	fmt.Print(out)
}

func getCacheAlgoliasList() []model.Algolia {
	res, _ := common.Exists(common.CACHE_ALGOLIA_JSON_PATH)
	cacheAlgiliasArray := []model.Algolia{}
	if res {
		jsonString := common.ReadFileString(common.CACHE_ALGOLIA_JSON_PATH)
		cacheAlgiliasArray = getAlgiliasJsonArray(jsonString)
		for _, algolias := range cacheAlgiliasArray {
			common.CacheAlgoliasMap[algolias.Uri] = algolias
		}
	}
	return cacheAlgiliasArray
}

func getAlgiliasJsonArray(jsonString string) []model.Algolia {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	var array []model.Algolia
	json.Unmarshal([]byte(jsonString), &array)

	return array
}

func getMd5Map() map[string]interface{} {
	md5Json := common.ReadFileString(common.MD5_ALGOLIA_JSON_PATH)
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	var md5Map map[string]interface{}
	json.Unmarshal([]byte(md5Json), &md5Map)
	return md5Map
}
