package mqtt

import (
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

// 包装 net.Conn
type Channel struct {
	// 短，长 id
	Id, LongId string

	// 原始连接
	Source net.Conn

	// 是否已被关闭
	Closed bool

	// 与连接相关联的 kv
	Attr map[string]interface{}

	// 进站数据处理
	Handler InboundHandler

	// 读写锁
	lock sync.RWMutex
}

func NewChannel(conn net.Conn) *Channel {
	c := &Channel{
		Id:      NewId(),
		LongId:  NewId(),
		Source:  conn,
		Closed:  false,
		Attr:    make(map[string]interface{}, 8),
		Handler: &DefaultInboundHandler{},
	}

	return c
}

// 写入数据
func (this *Channel) Write(data []byte) (int, error) {
	if len(data) > 0 && !this.Closed {
		return this.Source.Write(data)
	}

	return 0, nil
}

// 关闭连接，释放资源
func (this *Channel) Close() error {
	if !this.Closed {
		this.Closed = true

		return this.Source.Close()
	}

	return nil
}

// 读取数据
func (this *Channel) Read(buf []byte) (int, error) {
	return this.Source.Read(buf)
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

func NewId() string {

	return ""
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
