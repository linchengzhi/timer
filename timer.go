package timer

import (
	"time"
	"container/list"
)

const (
	interval = 1000 //间隔时间为1秒
	slot_num = 3600
)

type job func(...interface{})

type Task struct {
	key    int64 // 定时器唯一标识, 用于删除定时器
	delay  int64
	ruling int64 // 延迟时间
	circle int64 // 时间轮需要转动几圈
	num    int   //次数 -1为无限循环
	job    job
	data   []interface{}
}

type TimeWheel struct {
	interval      int64 //间隔
	slotNum       int64
	slots         []*list.List // 时间轮槽
	ticker        *time.Ticker
	timer         map[int64]int64 //定时器任务
	job           job
	currentRuling int   //当前刻度
	key           int64 //定时器标识
	addChan       chan *Task
	cancelChan    chan int64
	exitChan      chan bool
}

//新建并设置定时器 默认间隔为1秒
func New(j job, num ... int64) *TimeWheel {
	tw := new(TimeWheel)
	if len(num) > 0 {
		tw.interval = num[0]
	} else {
		tw.interval = interval
	}
	tw.slotNum = slot_num
	tw.addChan = make(chan *Task)
	tw.cancelChan = make(chan int64)
	tw.exitChan = make(chan bool)
	tw.ticker = time.NewTicker(time.Duration(tw.interval) * time.Millisecond)
	tw.slots = make([]*list.List, tw.slotNum)
	for i := int64(0); i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
	tw.timer = make(map[int64]int64)
	tw.job = j
	go tw.start()
	return tw
}

// Stop 停止时间轮
func (tw *TimeWheel) Exit() {
	tw.exitChan <- true
}

//参数 延迟时间(毫秒) 回调函数, 回调函数的参数
//若延迟时间小于间隔时间，则等于间隔时间
func (tw *TimeWheel) Add(delay int64, data ...interface{}) int64 {
	return tw.setTimer(0, 1, delay, nil, data...)
}

func (tw *TimeWheel) AddRepeat(num int, delay int64, data ...interface{}) int64 {
	return tw.setTimer(0, num, delay, nil, data...)
}

//优先到自定义函数
func (tw *TimeWheel) AddHasFunc(delay int64, j job, data ...interface{}) int64 {
	return tw.setTimer(0, 1, delay, j, data...)
}

func (tw *TimeWheel) AddRepeatHasFunc(num int, delay int64, j job, data ...interface{}) int64 {
	return tw.setTimer(0, num, delay, j, data...)
}

//移除定时任务
func (tw *TimeWheel) Cancel(key int64) {
	tw.cancelChan <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			if tw.currentRuling == int(tw.slotNum)-1 {
				tw.currentRuling = 0
			} else {
				tw.currentRuling++
			}
			tw.do()
		case task := <-tw.addChan:
			tw.add(task)
		case key := <-tw.cancelChan:
			tw.cancel(key)
		case <-tw.exitChan:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) setTimer(key int64, num int, delay int64, j job, data ...interface{}) int64 {
	if key == 0 {
		tw.key++
		key = tw.key
	}
	task := new(Task)
	if delay < tw.interval {
		delay = tw.interval
	}
	task.delay = delay
	task.key = key
	task.num = num
	task.job = j
	task.data = data
	tw.addChan <- task
	return task.key
}

func (tw *TimeWheel) add(task *Task) {
	task.circle = task.delay / (tw.slotNum * tw.interval)
	task.ruling = (int64(tw.currentRuling) + task.delay/tw.interval) % tw.slotNum
	tw.slots[task.ruling].PushBack(task)
	tw.timer[task.key] = task.ruling
}

func (tw *TimeWheel) cancel(key int64) {
	ruling, ok := tw.timer[key]
	if !ok {
		return
	}
	l := tw.slots[ruling]
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timer, key)
			l.Remove(e)
		}
	}
}

func (tw *TimeWheel) do() {
	var repeat = make([]*Task, 0)
	l := tw.slots[tw.currentRuling]
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		if task.job == nil {
			go tw.job(task.data)
		} else {
			go task.job(task.data)
		}
		next := e.Next()
		l.Remove(e)
		delete(tw.timer, task.key)
		e = next
		repeat = append(repeat, task)
	}
	go tw.repeat(repeat)
}

func (tw *TimeWheel) repeat(tasks[]*Task) {
	for _, task := range tasks {
		if task.num == 1 {
			return
		}
		num := task.num - 1
		if task.num == -1 {
			num = -1
		}
		tw.setTimer(task.key, num, task.delay, task.job, task.data...)
	}
}

