package traceroute

import (
	"sync"
	"time"
)

type notifyChannel struct {
	// 通知频道
	channel chan *TracerouteHop
	// 频道开关状态
	status bool

	lock *sync.RWMutex
}

func newNotifyChannel(size int) (object *notifyChannel) {
	object = &notifyChannel{
		channel: make(chan *TracerouteHop, size),
		status:  true,
		lock:    &sync.RWMutex{},
	}
	return object
}

func (self *notifyChannel) notify(hop *TracerouteHop) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if !self.status {
		return
	}

	select {
	case self.channel <- hop:
	case <-time.After(time.Millisecond):
	}
}

func (self *notifyChannel) close() {
	self.lock.Lock()
	defer self.lock.Unlock()

	if !self.status {
		return
	}

	self.status = false
	close(self.channel)
}
