package traceroute

import (
	"errors"
	"net"
	"sync"
)

type Options struct {
	// 最大跳数
	maxHops int
	// 发送包大小
	packetSize int
	// 超时时间
	timeout int64
	// 端口号
	port int
	// 第一跳位置
	firstHop int
	// 单跳最大重试次数
	retries int
	// 起始地址
	sourceAddress net.IP
	// 目标地址
	destinationAddress net.IP
	// 绑定的traceroute对象
	bindTrace *traceroute
	// 启动信号
	begin chan struct{}

	lock *sync.RWMutex
}

func NewOptions(destinationAddress string) (object *Options) {
	object = &Options{
		destinationAddress: net.ParseIP(destinationAddress),
		begin:              make(chan struct{}),
		lock:               &sync.RWMutex{},
	}
	return object
}

func (self *Options) bind(trace *traceroute) (bindErr error) {
	if self.lock == nil {
		self.lock = &sync.RWMutex{}
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	if self.bindTrace != nil {
		return errors.New("the current options is already in use")
	}

	self.bindTrace = trace
	close(self.begin)
	return nil
}

func (self *Options) unBind() {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.bindTrace = nil
	self.begin = make(chan struct{})
}

func (self *Options) Watch() <-chan *TracerouteHop {
	<-self.begin

	if self.bindTrace == nil {
		return nil
	}

	return self.bindTrace.increaseWatch()
}

func (self *Options) SetMaxHops(maxHops int) {
	self.maxHops = maxHops
}

func (repl Options) GetMaxHops() int {
	if repl.maxHops == 0 {
		return defaultMaxHops
	}

	return repl.maxHops
}

func (self *Options) SetPacketSize(packetSize int) {
	self.packetSize = packetSize
}

func (repl Options) GetPacketSize() int {
	if repl.packetSize == 0 {
		return defaultPacketSize
	}

	return repl.packetSize
}

func (self *Options) SetTimeout(timeout int64) {
	self.timeout = timeout
}

func (repl Options) GetTimeout() int64 {
	if repl.timeout == 0 {
		return defaultTimeout
	}

	return repl.timeout
}

func (self *Options) SetPort(port int) {
	self.port = port
}

func (repl Options) GetPort() int {
	if repl.port == 0 {
		return defaultPort
	}

	return repl.port
}

func (self *Options) SetFirstHop(firstHop int) {
	self.firstHop = firstHop
}

func (repl Options) GetFirstHop() int {
	if repl.firstHop == 0 {
		return defaultFirstHop
	}

	return repl.firstHop
}

func (self *Options) SetRetries(retries int) {
	self.retries = retries
}

func (repl Options) GetRetries() int {
	if repl.retries == 0 {
		return defaultRetries
	}

	return repl.retries
}

func (self *Options) SetSourceAddress(sourceAddress net.IP) {
	self.sourceAddress = sourceAddress
}

func (repl Options) GetSourceAddress() net.IP {
	if repl.sourceAddress == nil {
		sourceAddress, getErr := getInterfaceAddressIPv4()
		if getErr != nil {
			sourceAddress = net.ParseIP(localhost)
		}
		return sourceAddress
	}

	return repl.sourceAddress
}

func (self *Options) SetDestinationAddress(destinationAddress net.IP) {
	self.destinationAddress = destinationAddress
}

func (repl Options) GetDestinationAddress() net.IP {
	return repl.destinationAddress
}
