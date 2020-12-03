package store

import (
	"mqtt-go/src/message"
	"sync"
)

const PlaceHolder = true

var Store = &store{
	allTopics:    make(map[string]bool),
	topicClients: make(map[string]map[string]byte),
	clientTopics: make(map[string]map[string]byte),
}

// 存储服务
type store struct {
	// 全部主题, set 集合
	allTopics map[string]bool
	lock0     sync.RWMutex

	// topic(one) <--> clientId(many)
	topicClients map[string]map[string]byte

	// client(one) <--> topic(many)
	clientTopics map[string]map[string]byte
}

// 获取订阅指定 topic 的 client 集合
func (this *store) Search(topic string) []*message.ClientSub {
	this.lock0.RLock()
	defer this.lock0.RUnlock()

	m := this.topicClients[topic]
	if m == nil {
		return nil
	}
	clientIds := make([]*message.ClientSub, 0, len(m))
	for k, v := range m {
		clientIds = append(clientIds, &message.ClientSub{
			Qos:      v,
			ClientId: k,
		})
	}

	return clientIds
}

// 订阅
func (this *store) Subscribe(clientId string, topics ...*message.Topic) {
	this.lock0.Lock()
	defer this.lock0.Unlock()

	// client 订阅的 topic 集合
	if this.clientTopics[clientId] == nil {
		this.clientTopics[clientId] = make(map[string]byte)
	}
	clientTopics := this.clientTopics[clientId]

	for _, topic := range topics {
		clientTopics[topic.Name] = topic.Qos

		if this.topicClients[topic.Name] == nil {
			this.topicClients[topic.Name] = make(map[string]byte)
		}
		topicClientSet := this.topicClients[topic.Name]

		topicClientSet[clientId] = topic.Qos
	}
}

// 解除订阅
func (this *store) RemoveSub(clientId string, topics ...string) {
	this.lock0.Lock()
	defer this.lock0.Unlock()

	// client 订阅的 topic 集合
	clientTopics := this.clientTopics[clientId]
	if clientTopics == nil {
		// client 无订阅关系
		return
	}

	for _, topic := range topics {
		// 移除订阅
		delete(clientTopics, topic)

		// 主题客户端订阅集合关系移除
		topicClientSet := this.topicClients[topic]
		delete(topicClientSet, clientId)
	}
}

// 移除全部订阅
func (s *store) RemoveAllSub(clientId string) {
	s.lock0.Lock()
	defer s.lock0.Unlock()

	delete(s.clientTopics, clientId)
}
