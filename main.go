package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/json-iterator/go"
	"hugo-algolia-updater/constant1"
	"hugo-algolia-updater/po"
	"hugo-algolia-updater/utils"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	serviceName string
	version string
	buildTime string
)

var showVersion bool

func main() {
	fmt.Printf("%s %s %s @%s\n", serviceName, version, buildTime, runtime.Version())

	flag.BoolVar(&showVersion, "v", false, "show version and exit")
	flag.Parse()

	if showVersion {
		return
	}

	utils.InitJieba()

	startTime := time.Now().UnixNano() / 1e6

	//运行编译
	execHugoBuild()

	participlesStartTime := time.Now().UnixNano() / 1e6

	var articleList = getArticleList()

	//获取分词列表
	cacheAlgoliasList := getCacheAlgoliasList()
	var taskNum = 0
	var flag = true
	//有缓存时
	if len(cacheAlgoliasList) != 0 {
		exists, _ := utils.Exists(constant1.MD5_ALGOLIA_JSON_PATH)
		if exists {
			flag = false
			//有md5map
			constant1.Md5Map = po.NewConcurrentMap(getMd5Map())

			for _, article := range articleList {
				sss := article
				permalink := sss.Yaml.Permalink
				value := constant1.Md5Map.GetValue(permalink)
				oldMd5 := ""
				if value != nil {
					oldMd5 = value.(string)
				}
				compare := strings.Compare(oldMd5, sss.Md5Value)
				if compare != 0 {
					constant1.Queue.Push(sss)
					constant1.NeedArticleList = append(constant1.NeedArticleList, sss)
					taskNum++
				}
			}
		}
	}

	//没缓存时
	if flag {
		for _, article := range articleList {
			constant1.Queue.Push(article)
			constant1.NeedArticleList = append(constant1.NeedArticleList, article)
			taskNum++
		}

	}

	//创建WaitGroup（java中的countdown）
	constant1.WaitGroup.Add(taskNum)

	//创建线程池
	pool := new(utils.ThreadPool)
	pool.Init(runtime.NumCPU(), taskNum)

	//循环添加任务
	for i := 0; i < taskNum; i++ {
		pool.AddTask(ParticiplesAsynchronous)
	}
	pool.Start()

	//主线程阻塞
	constant1.WaitGroup.Wait()
	pool.Stop()
	fmt.Println("participles success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-participlesStartTime, 10) + " ms")

	//创建分词
	algoliaStartTime := time.Now().UnixNano() / 1e6
	for _, article := range constant1.NeedArticleList {
		constant1.CacheAlgoliasMap[article.Yaml.Permalink] = po.Algolia{Title: article.Yaml.Title}
		//cacheAlgoliasList = append(cacheAlgoliasList, po.Algolia{Title: value.Yaml.Title})
	}

	var objArray = []algoliasearch.Object{}
	for permalink, algolias := range constant1.CacheAlgoliasMap {

		value := constant1.ArticleMap.GetValue(permalink)
		var article *po.Article
		if value != nil {
			article = value.(*po.Article)
		} else {
			log.Printf("ArticleMap.GetValue failed, permalink=%v\n", permalink)
			continue
		}
		constant1.Md5Map.AddData(permalink, article.Md5Value)

		mapObj := utils.Struct2Map(article.Yaml)
		//fmt.Printf("Struct2Map %#v\n", mapObj)

		if article.Participles != nil {
			participlesArray := *article.Participles
			var buffer bytes.Buffer
			for _, str := range participlesArray {
				if constant1.NumberReg.Match([]byte(str)) {
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
		mapObj["objectID"] = article.Yaml.Permalink
		mapObj["uri"] = article.Yaml.Permalink

		objArray = append(objArray, mapObj)
	}
	fmt.Println("generate algolia index success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-algoliaStartTime, 10) + " ms")
	fmt.Println("generate algolia index num: ", constant1.Num)

	return
	uploadStartTime := time.Now().UnixNano() / 1e6
	//更新分词
	utils.UpdateAlgolia(objArray)
	fmt.Println("update algolia success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-uploadStartTime, 10) + " ms")
	saveStartTime := time.Now().UnixNano() / 1e6
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	algoliaBytes, _ := json.Marshal(objArray)
	md5Bytes, _ := json.Marshal(constant1.Md5Map.GetData())
	utils.WriteFile(constant1.ALGOLIA_COMPLIE_JSON_PATH, algoliaBytes)
	utils.WriteFile(constant1.CACHE_ALGOLIA_JSON_PATH, algoliaBytes)
	utils.WriteFile(constant1.MD5_ALGOLIA_JSON_PATH, md5Bytes)

	fmt.Println("save cache success: " + strconv.FormatInt((time.Now().UnixNano()/1e6)-saveStartTime, 10) + " ms")
	fmt.Println("total : " + strconv.FormatInt((time.Now().UnixNano()/1e6)-startTime, 10) + " ms")
}

func getArticleList() []*po.Article {
	hugoJsonFile := "public/index.json"
	c, err := ioutil.ReadFile(hugoJsonFile)
	if err != nil {
		panic(err)
	}
	var posts []*po.HugoJsonPost
	if err := json.Unmarshal(c, &posts); err != nil {
		panic(err)
	}

	var articleList []*po.Article
	taskNum := len(posts)
	//创建WaitGroup（java中的countdown）
	constant1.WaitGroup.Add(taskNum)

	//设置cpu并行数
	runtime.GOMAXPROCS(runtime.NumCPU())

	//创建线程池
	pool := new(utils.ThreadPool)
	pool.Init(runtime.NumCPU(), taskNum)

	for _, post := range posts {
		//log.Printf("post=%#v", post)
		post1 := post
		pool.AddTask(func() error {
			article := po.Article{Yaml: *post1, Content: post1.Contents, Md5Value: utils.Md5V(post1.Contents)}
			articleList = append(articleList, &article)
			constant1.ArticleMap.AddData(post1.Permalink, &article)
			constant1.WaitGroup.Done()
			return nil
		})
	}

	pool.Start()
	//主线程阻塞
	constant1.WaitGroup.Wait()
	pool.Stop()
	return articleList
}

//多线程分词
func ParticiplesAsynchronous() error {
	article := constant1.Queue.Pop().(*po.Article)
	content := article.Content
	mdConf := article.Yaml

	participles := utils.Participles(mdConf.Title, content)
	article.Participles = &participles
	fmt.Println("generate success: " + article.Yaml.Permalink)
	constant1.WaitGroup.Done()
	return nil
}

//执行编译
func execHugoBuild() {
	out, _ := utils.ExecShell("hugo", "--gc", "--minify", "--enableGitInfo")
	fmt.Print(out)
}

func getCacheAlgoliasList() []po.Algolia {
	var res, _ = utils.Exists(constant1.CACHE_ALGOLIA_JSON_PATH)
	cacheAlgiliasArray := []po.Algolia{}
	if res {
		jsonString := utils.ReadFileString(constant1.CACHE_ALGOLIA_JSON_PATH)
		cacheAlgiliasArray = getAlgiliasJsonArray(jsonString)
		for _, algolias := range cacheAlgiliasArray {
			constant1.CacheAlgoliasMap[algolias.Uri] = algolias
		}
	}
	return cacheAlgiliasArray
}

func getAlgiliasJsonArray(jsonString string) []po.Algolia {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var array []po.Algolia
	json.Unmarshal([]byte(jsonString), &array)

	return array
}

func getMd5Map() map[string]interface{} {
	md5Json := utils.ReadFileString(constant1.MD5_ALGOLIA_JSON_PATH)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var md5Map map[string]interface{}
	json.Unmarshal([]byte(md5Json), &md5Map)
	return md5Map
}
