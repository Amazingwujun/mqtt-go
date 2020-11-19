package channel

import (
	"log"
	"net"
	"os"
	"sync/atomic"
)

type Channel struct {

	// 短 id
	id string

	// 长 id
	longId string

	// 原始连接
	source net.Conn
}

func (this Channel) write() {

}

var (
	processId  int32
	sequenceId int32
)

func init() {

	processId = int32(os.Getpid())

	log.Printf("pid: %d\n", processId)

	atomic.AddInt32(&sequenceId, 10)

	log.Printf("sequenceId: %d\n", sequenceId)
}

func NewId() (string, error) {

	return "", nil
}

func machineId() (string, error) {
	// 优先抓取环境变量
	machineId := os.Getenv("mqttx.machineId")
	if machineId != "" {
		return machineId, nil
	}

	// 抓取机器 mac 地址
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, n := range interfaces {
		mac := n.HardwareAddr.String()
		if len(mac) == 0 {
			continue
		}

		return mac, nil
	}

	return "", nil
}
