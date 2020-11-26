package channel

import (
	"log"
	"math"
	"mqtt-go/src/codec"
	"mqtt-go/src/message"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

const CLIENT_ID = "CLIENT_ID"

var (
	idCounter uint64 = 0
)

// 字节池
var bytesPool = &sync.Pool{New: func() interface{} {
	return make([]byte, 512)
}}

// Channel 代表一个 TCP 连接及与连接相关的上下文, 处理逻辑中它绑定了固定的 goroutine
type Channel struct {
	// 短
	Id string

	// 原始连接
	origin net.Conn

	// 是否已被关闭
	Closed bool

	// 与连接相关联的 kv
	attr map[string]interface{}

	encoder func(msg *message.MqttMessage) []byte

	// []byte pool
	pool *sync.Pool

	// packageId
	packageId uint16

	// 输出流
	Out chan []byte

	// 关闭信号
	Stop chan struct{}

	// 读写锁
	lock sync.RWMutex
}

// 构建一个新的 Channel
func NewChannel(conn net.Conn) *Channel {
	c := &Channel{
		Id:        NewId(),
		origin:    conn,
		Closed:    false,
		attr:      make(map[string]interface{}, 8),
		Out:       make(chan []byte, 10),
		encoder:   codec.Encode,
		pool:      bytesPool,
		packageId: 0,
		Stop:      make(chan struct{}),
	}

	return c
}

// 返回下一个 packageId
func (this *Channel) NextPackageId() uint16 {
	if this.packageId == 0 {
		this.packageId++
	} else if math.MaxUint16 == this.packageId {
		this.packageId = 1
		return math.MaxUint16
	}
	return this.packageId
}

// 写入数据
func (this *Channel) Write(msg *message.MqttMessage) {
	this.Out <- this.encoder(msg)
}

// 直接写入数据
func (this *Channel) Write0(buf []byte) (int, error) {
	return this.origin.Write(buf)
}

// 关闭连接，释放资源
func (this *Channel) Close() error {
	if !this.Closed {
		this.Closed = true

		// 发送停止信号
		this.Stop <- struct{}{}

		return this.origin.Close()
	}

	return nil
}

func (this *Channel) Get() []byte {
	return this.pool.Get().([]byte)
}

func (this *Channel) Put(buf []byte) {
	this.pool.Put(buf)
}

func (this *Channel) HGet(k string) interface{} {
	return this.attr[k]
}

func (this *Channel) HPut(k string, v interface{}) {
	this.attr[k] = v
}

// 读取数据
func (this *Channel) Read(buf []byte) (int, error) {
	return this.origin.Read(buf)
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
	return strconv.FormatUint(atomic.AddUint64(&idCounter, 1), 10)
}

// mac 地址获取
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

// 返回与 Channel 关联的 clientId
func (this *Channel) ClientId() string {
	return this.attr[CLIENT_ID].(string)
}

// 返回与 Channel 关联的 clientId
func (this *Channel) SaveClientId(clientId string) {
	this.attr[CLIENT_ID] = clientId
}