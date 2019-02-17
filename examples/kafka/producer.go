package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	//ack等级,保证消息是否已存储到磁盘
	config.Producer.RequiredAcks = sarama.WaitForAll
	//随机的分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//等待成功响应
	config.Producer.Return.Successes = true
	//异步的生产者
	client, err := sarama.NewAsyncProducer([]string{"127.0.0.1:9092"}, config)
	// sarama.OffsetFetchRequest
	if err != nil {
		fmt.Println("producer close, err:", err)
		return
	}
	defer client.Close()
	go func() {
		for mqMsg := range client.Successes() {
			fmt.Printf("topic:%v,value:%v,partition:%v\n", mqMsg.Topic, mqMsg.Value, mqMsg.Partition)
		}
	}()
	go func() {
		for err := range client.Errors() {
			fmt.Println("err:", err)
		}
	}()

	for {
		msg := &sarama.ProducerMessage{}
		//消息主题,可以当做是事件名
		msg.Topic = "demo_msg"
		msg.Value = sarama.StringEncoder("lalala:" + strconv.Itoa(rand.Int()))

		client.Input() <- msg
		fmt.Println("send msg:", msg.Value)
		time.Sleep(2 * time.Second)
	}

}
