package domain

import (
	"sync/atomic"
	"unsafe"
)

type Endport struct {
	IP          string  `json:"ip"`
	Port        string  `json:"port"`
	ActiveSorce float64 `json:"-"`
	StaticSorce float64 `json:"-"`
	//平均值结果
	Stats *Stat `json:"-"`
	//保存一台机器的多次数据
	window *stateWindow `json:"-"`
}

func NewEndport(ip, port string) *Endport {
	ed := &Endport{
		IP:   ip,
		Port: port,
	}
	ed.window = newStateWindow()
	ed.Stats = ed.window.getStat()
	//匿名闭包函数
	go func() {
		for stat := range ed.window.statChan {
			ed.window.appendStat(stat)
			newStat := ed.window.getStat()
			//通过将两个指针转换为unsafe.Pointer类型并使用atomic.SwapPointer函数来实现指针的原子交换
			//如果有监听事件发生，newStat是一个临时变量，保存了最新的平均值。直接将这个临时变量和new交换值。
			atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(ed.Stats)), unsafe.Pointer(newStat))
		}
	}()
	return ed
}

func (ed *Endport) UpdateStat(s *Stat) {
	ed.window.statChan <- s
}

func (ed *Endport) CalculateScore(ctx *IpConfConext) {
	// 如果 stats 字段是空的，则直接使用上一次计算的结果，此次不更新
	if ed.Stats != nil {
		//更新两种分数的值
		ed.ActiveSorce = ed.Stats.CalculateActiveSorce()
		ed.StaticSorce = ed.Stats.CalculateStaticSorce()
	}
}
