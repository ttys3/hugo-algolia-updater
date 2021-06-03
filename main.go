package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ttys3/hugo-algolia-updater/config"

	"go.uber.org/zap"

	jsoniter "github.com/json-iterator/go"
	"github.com/ttys3/hugo-algolia-updater/common"
	"github.com/ttys3/hugo-algolia-updater/model"
)

var (
	serviceName string
	version     string
	buildTime   string
)

const DftConfigFile = "config.yaml"

var (
	showVersion        bool
	cleanGeneratedJson bool
	configFile         string
)

func main() {
	// init logger
	// https://github.com/uber-go/zap/issues/717#issuecomment-496612544
	// The default global logger used by zap.L() and zap.S() is a no-op logger.
	// To configure the global loggers, you must use ReplaceGlobals.
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()
	zap.L().Info("replaced zap's global loggers")

	// nolint: forbidigo
	fmt.Printf("%s %s %s @%s\n", serviceName, version, buildTime, runtime.Version())

	flag.BoolVar(&showVersion, "v", false, "show version and exit")
	flag.BoolVar(&cleanGeneratedJson, "clean", false, "clean generated json files")
	flag.StringVar(&configFile, "c", DftConfigFile, "config file path, if not specified, use config.yaml under current working dir")
	flag.Parse()

	if showVersion {
		return
	}

	if cleanGeneratedJson {
		toClean := []string{
			common.ALGOLIA_COMPLIE_JSON_PATH,
			common.CACHE_ALGOLIA_JSON_PATH,
			common.MD5_ALGOLIA_JSON_PATH,
			common.HUGO_INDEX_JSON_PATH,
		}
		for _, f := range toClean {
			if err := os.Remove(f); err == nil {
				zap.S().Infof("%s removed successfully", f)
			} else {
				if os.IsNotExist(err) {
					zap.S().Infof("%s not exists", f)
				} else {
					zap.S().Errorf("remove %s failed, err=%v", f, err)
				}
			}
		}
		return
	}

	if err := config.Cfg.Load(configFile); err != nil {
		zap.S().Fatal(err)
	}
	if err := config.Cfg.Validate(); err != nil {
		zap.S().Fatal(err)
	}

	zap.S().Infof("loaded config: %v", config.Cfg)

	jiebaShutdown := common.InitJieba(config.Cfg.AlgoliaUpdater.Segment.Dict.Path, config.Cfg.AlgoliaUpdater.Segment.Dict.StopPath)
	defer jiebaShutdown()

	startTime := time.Now().UnixNano() / 1e6

	// 运行编译
	if err := execHugoBuild(); err != nil {
		zap.S().Fatal(err)
	}

	segmentsStartTime := time.Now().UnixNano() / 1e6

	articleList := getArticleList()

	// 获取分词列表
	cacheAlgoliasList := getCacheAlgoliasList()
	taskNum := 0
	needSegFlag := true
	// 有缓存时
	if len(cacheAlgoliasList) != 0 {
		exists, _ := common.Exists(common.MD5_ALGOLIA_JSON_PATH)
		if exists {
			needSegFlag = false
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
	if needSegFlag {
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
		pool.AddTask(SegmentsAsynchronous)
	}
	pool.Start()

	// 主线程阻塞
	common.WaitGroup.Wait()
	pool.Stop()
	zap.S().Infof("segments success: %v ms", (time.Now().UnixNano()/1e6)-segmentsStartTime)

	// 创建分词
	algoliaStartTime := time.Now().UnixNano() / 1e6
	for _, article := range common.NeedArticleList {
		// TODO better handled insertion element to CacheAlgoliasMap, use full model.Algolia struct
		common.CacheAlgoliasMap[article.HugoJsonPost.Permalink] = &model.Algolia{Title: article.HugoJsonPost.Title}
	}

	var objArray []*model.Algolia
	for permalink := range common.CacheAlgoliasMap {

		value := common.ArticleMap.GetValue(permalink)
		var article *model.Article
		if value != nil {
			article = value.(*model.Article)
		} else {
			log.Printf("ArticleMap.GetValue failed, permalink=%v\n", permalink)
			continue
		}
		common.Md5Map.AddData(permalink, article.Md5Value)

		theURL := ""
		if u, err := url.Parse(article.HugoJsonPost.Permalink); err == nil {
			theURL = u.RequestURI()
		}
		algobj := model.Algolia{
			ObjectID:      article.HugoJsonPost.Permalink,
			Title:         article.HugoJsonPost.Title,
			Keywords:      article.HugoJsonPost.Tags,
			Description:   article.HugoJsonPost.Description,
			Content:       article.HugoJsonPost.Content,
			URL:           theURL,
			Lang:          "en",
			Origin:        "",
			Image:         "",
			DatePublished: article.HugoJsonPost.Date.Unix(),
			Subtitle:      "",
			Date:          article.HugoJsonPost.Date.String(),
			Author:        "",
			Tags:          article.HugoJsonPost.Tags,
			Categories:    article.HugoJsonPost.Categories,
		}
		if len(article.HugoJsonPost.Images) > 0 {
			algobj.Image = article.HugoJsonPost.Images[0]
		}

		if article.Segments != nil {
			segmentsArray := *article.Segments
			var buffer bytes.Buffer
			for _, str := range segmentsArray {
				if common.NumberReg.Match([]byte(str)) {
					continue
				}
				buffer.WriteString(str)
				buffer.WriteString(" ")
			}
			join := buffer.String()
			algobj.Content = join
		}
		objArray = append(objArray, &algobj)
	}
	zap.S().Infof("generate algolia index success: %v ms", (time.Now().UnixNano()/1e6)-algoliaStartTime)
	zap.S().Infof("generate algolia index num: %v", common.Num)

	uploadStartTime := time.Now().UnixNano() / 1e6
	// 更新分词
	if err := common.UpdateAlgolia(config.Cfg.AlgoliaUpdater.Algolia.Index,
		config.Cfg.AlgoliaUpdater.Algolia.AppID,
		config.Cfg.AlgoliaUpdater.Algolia.AdminKey,
		objArray); err == nil {
		zap.S().Infof("update algolia index %s success: %v ms, %v objects",
			config.Cfg.AlgoliaUpdater.Algolia.Index, (time.Now().UnixNano()/1e6)-uploadStartTime, len(objArray))
	} else {
		zap.S().Fatal(err)
	}

	saveStartTime := time.Now().UnixNano() / 1e6
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	algoliaBytes, _ := json.Marshal(objArray)
	md5Bytes, _ := json.Marshal(common.Md5Map.GetData())
	common.WriteFile(common.ALGOLIA_COMPLIE_JSON_PATH, algoliaBytes)
	common.WriteFile(common.CACHE_ALGOLIA_JSON_PATH, algoliaBytes)
	common.WriteFile(common.MD5_ALGOLIA_JSON_PATH, md5Bytes)

	zap.S().Infof("save cache success: %v ms", (time.Now().UnixNano()/1e6)-saveStartTime)
	zap.S().Infof("total : %v ms", (time.Now().UnixNano()/1e6)-startTime)
}

func getArticleList() []*model.Article {
	hugoJsonFile := common.HUGO_INDEX_JSON_PATH
	c, err := ioutil.ReadFile(hugoJsonFile)
	if err != nil {
		panic(err)
	}
	var posts []*model.HugoJsonPost
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(c, &posts); err != nil {
		panic(err)
	}

	var articleList []*model.Article
	taskNum := len(posts)
	// 创建WaitGroup（java中的countdown）
	common.WaitGroup.Add(taskNum)

	// 创建线程池
	pool := new(common.ThreadPool)
	pool.Init(runtime.NumCPU(), taskNum)

	for _, post := range posts {
		// log.Printf("post=%#v", post)
		if post.Content == "" {
			pool.AddTask(func() error {
				// we need this since the pool has been inited with fixed number of tasks
				common.WaitGroup.Done()
				return nil
			})
			continue
		}
		post1 := post
		pool.AddTask(func() error {
			article := model.Article{HugoJsonPost: *post1, Md5Value: common.Md5V(post1.Content)}
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
func SegmentsAsynchronous() error {
	article := common.Queue.Pop().(*model.Article)
	content := article.HugoJsonPost.Content
	title := article.HugoJsonPost.Title

	segments := common.DoSegment(title, content)
	article.Segments = &segments
	zap.S().Infof("generate success: " + article.HugoJsonPost.Permalink)
	common.WaitGroup.Done()
	return nil
}

// 执行编译
func execHugoBuild() error {
	out, err := common.ExecShell("hugo", "--gc", "--enableGitInfo")
	if err != nil {
		return err
	}
	zap.S().Info(out)
	return nil
}

func getCacheAlgoliasList() []*model.Algolia {
	res, _ := common.Exists(common.CACHE_ALGOLIA_JSON_PATH)
	cacheAlgiliasArray := []*model.Algolia{}
	if res {
		jsonString := common.ReadFileString(common.CACHE_ALGOLIA_JSON_PATH)
		cacheAlgiliasArray = getAlgiliasJsonArray(jsonString)
		for _, algolias := range cacheAlgiliasArray {
			algolias1 := algolias
			common.CacheAlgoliasMap[algolias.ObjectID] = algolias1
		}
	}
	return cacheAlgiliasArray
}

func getAlgiliasJsonArray(jsonString string) []*model.Algolia {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	var array []*model.Algolia
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
