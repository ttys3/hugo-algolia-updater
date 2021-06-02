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
	CacheAlgoliasMap = map[string]*model.Algolia{}
	Md5Map           = NewConcurrentMap(make(map[string]interface{}))
	NeedArticleList  = []*model.Article{}
	ArticleMap       = NewConcurrentMap(make(map[string]interface{}))
	StopArray        = []string{}
	NumberReg        = regexp.MustCompile(`\d+|\d+\.+\d+`)
)

const N int = 10

type QueueNode struct {
	data *list.List
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
		// nolint: forbidigo
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
