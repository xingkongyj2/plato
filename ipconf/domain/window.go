package domain

const (
	windowSize = 5
)

type stateWindow struct {
	//使用一个滑动窗口，保存多个state
	stateQueue []*Stat
	statChan   chan *Stat
	sumStat    *Stat
	idx        int64
}

func newStateWindow() *stateWindow {
	return &stateWindow{
		stateQueue: make([]*Stat, windowSize),
		statChan:   make(chan *Stat),
		sumStat:    &Stat{},
	}
}

func (sw *stateWindow) getStat() *Stat {
	res := sw.sumStat.Clone()
	res.Avg(windowSize)
	return res
}

func (sw *stateWindow) appendStat(s *Stat) {
	// 减去即将被删除的state
	sw.sumStat.Sub(sw.stateQueue[sw.idx%windowSize])
	// 更新最新的stat
	sw.stateQueue[sw.idx%windowSize] = s
	// 计算最新的窗口和
	sw.sumStat.Add(s)
	sw.idx++
}
