package traceroute

import (
	"errors"
	"fmt"
	"net"
)

// getInterfaceAddressIPv4 从网卡中获取服务器公网ipv4地址
func getInterfaceAddressIPv4() (outer net.IP, getErr error) {
	interfaceAddrs, err := net.InterfaceAddrs()

	if err != nil {
		return nil, fmt.Errorf("get interface addresses failed: %v ", err)
	}

	for _, address := range interfaceAddrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if len(ipNet.IP.To4()) == net.IPv4len {
				return ipNet.IP, nil
			}
		}
	}

	return nil, errors.New("can not find the client ip address")
}

func ipv4ToByte(origin net.IP) (result [4]byte) {
	if origin != nil && len(origin.To4()) == net.IPv4len {
		copy(result[:], origin.To4())
	}
	return
}

func byteToIPv4(origin [4]byte) (result net.IP) {
	result = net.IPv4(origin[0], origin[1], origin[2], origin[3])
	return
}
