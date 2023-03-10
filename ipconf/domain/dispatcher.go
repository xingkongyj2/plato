package domain

import (
	"sort"
	"sync"

	"github.com/hardcore-os/plato/ipconf/source"
)

type Dispatcher struct {
	//候选表
	candidateTable map[string]*Endport
	sync.RWMutex
}

// dp是一个调度器，是一个全局的，要注意并发问题。
var dp *Dispatcher

func Init() {
	//实例化调度器
	dp = &Dispatcher{}
	//创建候选表
	dp.candidateTable = make(map[string]*Endport)
	go func() {
		//source从etcd监听到数据后，回EventChan，这里监听EventChan
		for event := range source.EventChan() {
			switch event.Type {
			case source.AddNodeEvent:
				dp.addNode(event)
			case source.DelNodeEvent:
				dp.delNode(event)
			}
		}
	}()
}
func Dispatch(ctx *IpConfConext) []*Endport {
	// Step1（核心）: 获得候选endport：通过etcd得到
	//关键在于：candidateTable是怎么得到数据的，是通过异步的。在source初始化的时候就设置了
	eds := dp.getCandidateEndport(ctx)
	// Step2: 逐一计算得分
	for _, ed := range eds {
		ed.CalculateScore(ctx)
	}
	// Step3: 全局排序，动静结合的排序策略。
	sort.Slice(eds, func(i, j int) bool {
		// 优先基于活跃分数进行排序
		if eds[i].ActiveSorce > eds[j].ActiveSorce {
			return true
		}
		// 如果活跃分数相同，则使用静态分数排序
		if eds[i].ActiveSorce == eds[j].ActiveSorce {
			if eds[i].StaticSorce > eds[j].StaticSorce {
				return true
			}
			return false
		}
		return false
	})
	return eds
}

func (dp *Dispatcher) getCandidateEndport(ctx *IpConfConext) []*Endport {
	//这里加了读锁，可以无锁优化
	dp.RLock()
	defer dp.RUnlock()
	candidateList := make([]*Endport, 0, len(dp.candidateTable))
	//因为初始化的时候已经设置了candidateTable的数据监听，所有只要监听到数据，candidateTable就会自动更新
	//所以只要gateway注册一个服务，并且添加数据，这里就能有数据了
	//拿到候选值，复制到candidateList，并且返回出去
	for _, ed := range dp.candidateTable {
		candidateList = append(candidateList, ed)
	}
	//返回的数据是复制得到的，不会有并发的问题
	return candidateList
}
func (dp *Dispatcher) delNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	delete(dp.candidateTable, event.Key())
}
func (dp *Dispatcher) addNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	var (
		ed *Endport
		ok bool
	)
	//拿到监听事件的数据后，有两个异步更新操作：
	//    1.在候选表中使用IP+Port当成事件存在的唯一标识
	//    2.创建一个Endport对象，并且设置一个异步监听，监听state
	//简单理解：如果发现一个新增的event数据，就用IP+Port当成key，新建一个Endport对象当成value。同时开一个协程，异步的监听这个Endport中的state
	//注意：是每一个Endport都会有一个协程来监听自己数据中的state
	if ed, ok = dp.candidateTable[event.Key()]; !ok { // 不存在
		ed = NewEndport(event.IP, event.Port)
		dp.candidateTable[event.Key()] = ed
	}
	//更新统计数据
	ed.UpdateStat(&Stat{
		ConnectNum:   event.ConnectNum,
		MessageBytes: event.MessageBytes,
	})

}
