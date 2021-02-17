package base

import (
	"sync"
	"time"
)

// title: 主进度标题，sub：子进度标题
// no: 当前进度号（从1起），count：进度数量
// step:当前进度条进度，total：当前进度条总数
// arg：附加信息
// todo 暂且不支持取消功能（返回值：是否被取消；true：代表取消，false：代表继续）
type ProgressFunc func(title, sub string, no, count int, step, total int64, cost, estimate int) bool

// 进度条 Progress Bar
type Progress struct {
	title string
	bars  []*Bar
	fun   ProgressFunc
}

func NewProgress(title string, fun ProgressFunc) *Progress {
	p := &Progress{
		title: title,
		fun:   fun,
	}
	return p
}

// 添加子进度条
func (p *Progress) NewBar(sub string, total int64) *Bar {
	bar := &Bar{
		no:    len(p.bars),
		sub:   sub,
		total: total,
		start: time.Now().UnixNano(),
	}
	bar.root = p
	p.bars = append(p.bars, bar)
	return bar
}

// =========================================================== //

type Bar struct {
	lock sync.Mutex
	root *Progress

	no  int
	sub string // 子标题

	current int64 // 当前进度个数
	total   int64 // 总个数
	rate    int   // 当前进度百分比
	speed   int   // 从开始到现在的平均速度（个数/已花费时间）

	start    int64 // 开始时间点（单位：纳秒）
	cost     int   // 已花费时间（单位：秒）
	estimate int   // 预计剩余时间（单位：秒）
}

func (b *Bar) Add(n int64) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.current += n

	lastRate := b.rate
	b.rate = int(b.current * 100 / b.total) // 当前进度百分比（%）

	lastCost := b.cost
	now := time.Now().UnixNano()
	// 这个单位本应是纳秒（一秒等于10e9纳秒），但不知道为什么总是太慢，故而先乘以10
	cost := 10 * (now - b.start)
	b.cost = int(cost / 10e9) // 当前已花费时间（秒）
	// 满足以下条件之一时，进行刷新:1) ratio+1；2）时间过了一秒；3）current到了整百
	if b.current%100 == 0 || b.rate > lastRate || b.cost > lastCost {
		return b.refresh(cost)
	}
	return false
}

func (b *Bar) refresh(cost int64) bool {
	if cost != 0 {
		b.speed = int(int64(b.current) * 10e9 / cost) // 速度为 当前个数/已花费时间
		if b.speed != 0 {
			b.estimate = int((b.total - b.current) * cost / b.current / 10e9)
		}
	}
	if b.root.fun != nil {
		return b.root.fun(b.root.title, b.sub, b.no+1, len(b.root.bars), b.current, b.total, b.cost, b.estimate)
	}
	return false
}
