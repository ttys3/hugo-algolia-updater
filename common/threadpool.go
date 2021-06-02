package common

import "go.uber.org/zap"

type ThreadPool struct {
	Queue  chan func() error
	Number int
	Total  int

	result         chan error
	finishCallback func()
}

// 初始化
func (p *ThreadPool) Init(number int, total int) {
	p.Queue = make(chan func() error, total)
	p.Number = number
	p.Total = total
	p.result = make(chan error, total)
}

// 开始执行
func (p *ThreadPool) Start() {
	// 开启Number个goroutine
	for i := 0; i < p.Number; i++ {
		go func() {
			for {
				task, ok := <-p.Queue
				if !ok {
					break
				}

				err := task()
				p.result <- err
			}
		}()
	}

	// 获得每个work的执行结果
	for j := 0; j < p.Total; j++ {
		res, ok := <-p.result
		if !ok {
			break
		}

		if res != nil {
			zap.S().Info(res)
		}
	}

	// 所有任务都执行完成，回调函数
	if p.finishCallback != nil {
		p.finishCallback()
	}
}

// 停止
func (p *ThreadPool) Stop() {
	close(p.Queue)
	close(p.result)
}

// 添加任务
func (p *ThreadPool) AddTask(task func() error) {
	p.Queue <- task
}

// 设置结束回调
func (p *ThreadPool) SetFinishCallback(callback func()) {
	p.finishCallback = callback
}
