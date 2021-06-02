package common

import (
	"container/list"
	"fmt"
	"regexp"
	"sync"

	"github.com/ttys3/hugo-algolia-updater/model"
)

var (
	WaitGroup        = sync.WaitGroup{}
	Queue            = NewQueue()
	AlgoliasMap      = map[string]model.Algolia{}
	CacheAlgoliasMap = map[string]model.Algolia{}
	Md5Map           = NewConcurrentMap(make(map[string]interface{}))
	NeedArticleList  = []*model.Article{}
	NeedAlgoliasList = []*model.Algolia{}
	ArticleMap       = NewConcurrentMap(make(map[string]interface{}))
	StopArray        = []string{}
	HtmlReg, _       = regexp.Compile("<.{0,200}?>")
	PointReg, _      = regexp.Compile("\n|\t|\r")
	NumberReg, _     = regexp.Compile("[0-9]+|[0-9]+\\.+[0-9]+")
)

const N int = 10

type QueueNode struct {
	digits1 [N]int
	digits2 [N]int
	sflag   bool
	data    *list.List
}

var lock sync.Mutex

func NewQueue() *QueueNode {
	q := new(QueueNode)
	q.data = list.New()
	return q
}

func (q *QueueNode) Push(v interface{}) {
	defer lock.Unlock()
	lock.Lock()
	q.data.PushFront(v)
}

func (q *QueueNode) Dump() {
	for iter := q.data.Back(); iter != nil; iter = iter.Prev() {
		fmt.Println("item:", iter.Value)
	}
}

func (q *QueueNode) Pop() interface{} {
	defer lock.Unlock()
	lock.Lock()
	iter := q.data.Back()
	v := iter.Value
	q.data.Remove(iter)
	return v
}

func (q *QueueNode) Size() int {
	return q.data.Len()
}
