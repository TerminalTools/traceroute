package traceroute

import (
	"context"
	"net"
	"sync"
	"time"
)

type TracerouteHop struct {
	// 路由追踪结果
	address string
	// 路由追踪域名
	route string
	// 当前跳数
	ttl int
	// 本次追踪是否成功
	success bool
	// 耗时
	elapsed time.Duration
	// 域名查找进度
	lookupFinish bool
}

func (repl TracerouteHop) GetAddress() string {
	return repl.address
}

func (repl TracerouteHop) GetRoute() string {
	return repl.route
}

func (repl TracerouteHop) GetTTL() int {
	return repl.ttl
}

func (repl TracerouteHop) GetSuccess() bool {
	return repl.success
}

func (repl TracerouteHop) GetElapsed() time.Duration {
	return repl.elapsed
}

func (repl TracerouteHop) Finish() bool {
	return repl.lookupFinish
}

func (self *TracerouteHop) setFinish() {
	self.lookupFinish = true
}

func (self *TracerouteHop) LookupAddr() {
	defer self.setFinish()

	var (
		withTimeout, cancel = context.WithTimeout(context.Background(), lookupTimeout) // 设置超时
		currentHost, newErr = net.DefaultResolver.LookupAddr(withTimeout, self.address)
	)
	cancel()
	if newErr != nil {
		return
	}

	if len(currentHost) > 0 {
		self.route = currentHost[0]
	}
}

// WaitLookup 等待路由信息查询完域名信息
func WaitLookup(hops []*TracerouteHop) {
	wait := &sync.WaitGroup{}

	for _, hop := range hops {
		wait.Add(1)
		go waitLookup(hop, wait)
	}

	wait.Wait()
}

func waitLookup(hop *TracerouteHop, wait *sync.WaitGroup) {
	defer wait.Done()
	if hop == nil {
		return
	}

	for waitCount := 0; waitCount < lookupWaitMaxCount; waitCount++ {
		if hop.lookupFinish {
			return
		}
		<-time.NewTimer(lookupRefreshTime).C
	}
}
