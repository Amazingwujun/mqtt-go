package main

import (
	"fmt"
	"mqtt-go/mqtt"
	"testing"
)

type Demo struct {
	name string
}

func TestDemo(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("你是猪吗? %s", err)
		}
	}()

	buf := make([]byte, 15, 500)
	fmt.Printf("buf 长度:%d 容量:%d\n", len(buf), cap(buf))

	buf[5] = 0xff

	fmt.Printf("%v\n", buf)

	bytes := buf[4:10]
	fmt.Printf("%v\n", bytes)
	bytes[1] = 0xf1

	fmt.Printf("%v\n", buf)
}

func TestDemo1(t *testing.T) {
	s := []byte{0, 4, 'a', 'b', 'c', 'd'}
	fmt.Printf("%v\n", s)
	mqttString := mqtt.DecodeMqttString(s)
	fmt.Printf(mqttString)
}
