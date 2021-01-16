// 和操作系统等内核层面打交道的功能组
package base

import (
	"sync"
)

type GoMax struct {
	max int
	ch  chan struct{}
	wg  sync.WaitGroup
	// lock sync.Mutex // 并发写的时候，用来上锁
}

func (this *GoMax) Init(max int) {
	this.max = max
	this.ch = make(chan struct{}, this.max)
}

func (this *GoMax) Add() {
	this.wg.Add(1)
	// this.lock.Lock()
	this.ch <- struct{}{}
	// this.lock.Unlock()
}

func (this *GoMax) Done() {
	this.wg.Done()
	// this.lock.Lock()
	<-this.ch
	// this.lock.Unlock()
}

func (this *GoMax) Wait() {
	this.wg.Wait()
	close(this.ch)
}
