package traceroute

import (
	"errors"
	"syscall"
	"time"
)

type traceroute struct {
	// 端口号
	port int
	// 当前跳数
	ttl int
	// 最大跳数
	maxHops int
	// 发送包大小
	packetSize int
	// 单跳最大重试次数
	retries int
	// 超时
	timeval *syscall.Timeval
	// 起始地址
	sourceAddress [4]byte
	// 目标地址
	destinationAddress [4]byte
	// 路由追踪信息数组
	hops []*TracerouteHop
	// 订阅频道
	channels []*notifyChannel
	// 本次追踪已存在的路由
	existed map[string]bool
	// 单跳重试次数
	thisRetry int
	// 已经连续重试次数
	continuousRetry int
}

func Traceroute(options *Options) (hops []*TracerouteHop, tracerouteErr error) {
	if options == nil {
		return nil, errors.New("options cannot be empty")
	}

	if options.GetDestinationAddress() == nil {
		return nil, errors.New("destination address cannot be empty")
	}

	var (
		timeout = options.GetTimeout()
		timeval = syscall.NsecToTimeval(timeout)
	)

	var (
		trace = &traceroute{
			maxHops:            options.GetMaxHops(),
			packetSize:         options.GetPacketSize(),
			port:               options.GetPort(),
			ttl:                options.GetFirstHop(),
			retries:            options.GetRetries(),
			timeval:            &timeval,
			sourceAddress:      ipv4ToByte(options.GetSourceAddress()),
			destinationAddress: ipv4ToByte(options.GetDestinationAddress()),
			hops:               make([]*TracerouteHop, 0, options.GetMaxHops()),
			channels:           make([]*notifyChannel, 0),
			existed:            make(map[string]bool),
			thisRetry:          0,
			continuousRetry:    0,
		}
	)

	if bindErr := options.bind(trace); bindErr != nil {
		return nil, bindErr
	}

	defer options.unBind()

	return trace.hops, trace.do()
}

func (repl traceroute) isDestination(tracerouteHop *TracerouteHop) bool {
	if tracerouteHop == nil {
		return false
	}

	if byteToIPv4(repl.destinationAddress).String() == tracerouteHop.address {
		return true
	}

	return false
}

func (self *traceroute) increaseWatch() <-chan *TracerouteHop {
	channel := newNotifyChannel(self.maxHops)
	self.channels = append(self.channels, channel)

	return channel.channel
}

func (self *traceroute) notify(hop *TracerouteHop) {
	for _, notifyChannel := range self.channels {
		go notifyChannel.notify(hop)
	}
}

func (self *traceroute) done() {
	var (
		channelNumber = len(self.channels)
		wait          = time.NewTimer(time.Millisecond * time.Duration(channelNumber))
	)

	<-wait.C // 等待一会再关闭通道

	for _, notifyChannel := range self.channels {
		notifyChannel.close()
	}
}

func (self *traceroute) do() (doErr error) {
	defer self.done()

	for {

		tracerouteHop, tracerouteErr := self.traceroute()

		if tracerouteErr != nil {
			return tracerouteErr
		}

		if tracerouteHop == nil {
			if self.thisRetry < self.retries {
				self.thisRetry++
				continue
			}
			self.continuousRetry++
			tracerouteHop = &TracerouteHop{success: false, ttl: self.ttl, lookupFinish: true}
		} else {
			self.continuousRetry = 0
			if self.existed[tracerouteHop.address] {
				return
			}

			self.existed[tracerouteHop.address] = true
		}

		self.notify(tracerouteHop)
		self.hops = append(self.hops, tracerouteHop)
		self.thisRetry = 0
		self.ttl++

		if self.isDestination(tracerouteHop) {
			return nil
		}

		if self.ttl >= self.maxHops || self.continuousRetry >= defaultMaxContinuousRetry { // 路由追踪失败, 自定义最后一跳信息
			tracerouteHop = &TracerouteHop{
				address:      byteToIPv4(self.destinationAddress).String(),
				success:      true,
				ttl:          self.ttl,
				lookupFinish: true,
			}
			<-time.NewTimer(time.Millisecond).C // 等待1毫秒再添加最后一跳信息
			self.notify(tracerouteHop)
			self.hops = append(self.hops, tracerouteHop)
			return nil
		}
	}
}

func (repl traceroute) traceroute() (hop *TracerouteHop, tracerouteErr error) {
	tracerouteStartTime := time.Now()

	// Set up the socket to receive inbound packets
	recvSocket, newErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if newErr != nil {
		return nil, newErr
	}
	defer syscall.Close(recvSocket)

	// Set up the socket to send packets out.
	sendSocket, newErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if newErr != nil {
		return nil, newErr
	}
	defer syscall.Close(sendSocket)

	// This sets the current hop TTL
	setsockoptIntErr := syscall.SetsockoptInt(sendSocket, socketLevel, syscall.IP_TTL, repl.ttl)
	if setsockoptIntErr != nil {
		return nil, setsockoptIntErr
	}
	// This sets the timeout to wait for a response from the remote host

	setsockoptTimevalErr := syscall.SetsockoptTimeval(recvSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, repl.timeval)
	if setsockoptTimevalErr != nil {
		return nil, setsockoptTimevalErr
	}

	// Bind to the local socket to listen for ICMP packets
	bindErr := syscall.Bind(recvSocket, &syscall.SockaddrInet4{Port: repl.port, Addr: repl.sourceAddress})
	if bindErr != nil {
		return nil, bindErr
	}

	// Send a single null byte UDP packet
	sendtoErr := syscall.Sendto(sendSocket, []byte{socketPayload}, socketFlags, &syscall.SockaddrInet4{Port: repl.port, Addr: repl.destinationAddress})
	if sendtoErr != nil {
		return nil, sendtoErr
	}

	_, recvFrom, recvErr := syscall.Recvfrom(recvSocket, make([]byte, repl.packetSize), socketFlags)
	if recvErr != nil {
		return nil, nil
	}

	currentAddress, typeOK := recvFrom.(*syscall.SockaddrInet4)
	if !typeOK {
		return nil, nil
	}

	hop = &TracerouteHop{
		success: true,
		address: byteToIPv4(currentAddress.Addr).String(),
		ttl:     repl.ttl,
		elapsed: time.Since(tracerouteStartTime),
	}
	go hop.LookupAddr()

	return hop, nil
}
