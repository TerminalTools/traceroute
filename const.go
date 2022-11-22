package traceroute

import (
	"time"
)

// default
const (
	defaultMaxHops            = 64                            // 最大跳数
	defaultPacketSize         = 52                            // 包大小
	defaultTimeout            = int64(time.Millisecond) * 500 // 超时
	defaultPort               = 33434                         // 端口
	defaultFirstHop           = 1                             // 起始追踪位置
	defaultRetries            = 3                             // 单跳重试次数
	defaultMaxContinuousRetry = 10                            // 最大连续重试次数
)

// socket
const (
	socketLevel   = 0x0
	socketPayload = 0x0
	socketFlags   = 0
	localhost     = "127.0.0.1"
)

// lookup
const (
	lookupTimeout      = time.Second * 5        // 域名查找超时时间
	lookupRefreshTime  = time.Millisecond * 500 // 查找状态刷新时间
	lookupWaitMaxCount = 10                     // 最大等待次数
)
