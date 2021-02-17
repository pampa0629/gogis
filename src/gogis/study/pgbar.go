package main

// #include "Windows.h"
import "C"

import (
	"fmt"
	"sync"
	"time"
)

func init033() {
	var mode C.DWORD
	handle := C.GetStdHandle(C.STD_OUTPUT_HANDLE)
	C.GetConsoleMode(handle, &mode)
	// Enable Virtual Terminal Processing by adding the flag to the current mode
	C.SetConsoleMode(handle, mode|0x0004)
}

// 进程条 Progress Bar
type Pgbar struct {
	title   string
	maxLine int
}

func NewPgbar(title string) *Pgbar {
	init033()
	p := &Pgbar{
		title:   title,
		maxLine: gMaxLine,
	}
	gMaxLine++
	printf(gMaxLine, "%s", title)
	return p
}

// 添加子进程条
func (p *Pgbar) NewSubbar(prefix string, total int64) *Bar {
	gMaxLine++
	return newBar(gMaxLine, prefix, total)
}

func (p *Pgbar) End() {
	move(gMaxLine)
}

// =========================================================== //

type Bar struct {
	lock   sync.Mutex
	line   int    // 当前光标行号
	prefix string // 子标题
	width  int    // 整个进度条的总宽度，固定为100

	current int64 // 当前进度个数
	total   int64 // 总个数
	rate    int   // 当前进度百分比
	speed   int   // 从开始到现在的平均速度（个数/已花费时间）

	start    int64 // 开始时间点（单位：纳秒）
	cost     int   // 已花费时间（单位：秒）
	estimate int   // 预计剩余时间（单位：秒）
}

var (
	bar1 string
	bar2 string
)

func initBar(width int) {
	for i := 0; i < width; i++ {
		bar1 += "="
		bar2 += "-"
	}
}

func newBar(line int, prefix string, total int64) *Bar {
	if total <= 0 {
		return nil
	}

	if line <= 0 {
		gMaxLine++
		line = gMaxLine
	}

	bar := &Bar{
		line:   line,
		prefix: prefix,
		total:  total,
		width:  100,
		start:  time.Now().UnixNano(),
	}

	initBar(bar.width)
	return bar
}

func (b *Bar) Add(n int64) {
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
		b.refresh(cost)
	}
}

func (b *Bar) refresh(cost int64) {
	if cost != 0 {
		b.speed = int(int64(b.current) * 10e9 / cost) // 速度为 当前个数/已花费时间
		if b.speed != 0 {
			b.estimate = int((b.total - b.current) * cost / b.current / 10e9)
		}
	}
	printf(b.line, "\r%s", b.barMsg())
}

func (b *Bar) barMsg() string {
	prefix := fmt.Sprintf("%s", b.prefix)
	rate := fmt.Sprintf("%3d%%", b.rate)
	speed := fmt.Sprintf("%dps", b.speed)
	cost := b.timeFmt(b.cost)
	estimate := b.timeFmt(b.estimate)
	ct := fmt.Sprintf(" (%d/%d)", b.current, b.total)
	barLen := b.width - len(prefix) - len(rate) - len(speed) - len(cost) - len(estimate) - len(ct) - 10
	bar1Len := barLen * b.rate / 100
	bar2Len := barLen - bar1Len

	realBar1 := bar1[:bar1Len]
	var realBar2 string
	if bar2Len > 0 {
		realBar2 = ">" + bar2[:bar2Len-1]
	}

	msg := fmt.Sprintf(`%s %s%s [%s%s] %s %s in: %s`, prefix, rate, ct, realBar1, realBar2, speed, cost, estimate)
	switch {
	case b.rate < 40:
		return "\033[0;33m" + msg + "\033[0m" // 棕色
	case b.rate >= 40 && b.rate < 100:
		return "\033[0;36m" + msg + "\033[0m" // 青色
	default:
		return "\033[0;32m" + msg + "\033[0m" // 绿色
	}
}

func (b *Bar) timeFmt(cost int) string {
	var h, m, s int
	h = cost / 3600
	m = (cost - h*3600) / 60
	s = cost - h*3600 - m*60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// =========================================================== //

var (
	mu           sync.Mutex
	gSrcLine     = 0 //起点行
	gCurrentLine = 0 //当前行
	gMaxLine     = 0 //最大行
)

func move(line int) {
	fmt.Printf("\033[%dA\033[%dB", gCurrentLine, line)
	gCurrentLine = line
}

func printf(line int, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	move(line)
	fmt.Printf("\r"+format, args...)
	move(gMaxLine)
}
