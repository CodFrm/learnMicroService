package main

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
)

type exampleConsumerGroupHandler struct{}

func (exampleConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (exampleConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h exampleConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	//这里的消费者是一个消费组的demo
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Consumer.Return.Errors = true
	client, err := sarama.NewClient([]string{"127.0.0.1:9092"}, config)
	if err != nil {
		fmt.Println("Failed to start client: %v", err)
		return
	}
	defer client.Close()
	group, err := sarama.NewConsumerGroupFromClient("demo", client) //demo群组
	if err != nil {
		fmt.Println("group err: %v", err)
		return
	}
	ctx := context.Background()
	for {
		topics := []string{"demo_msg"}
		handler := exampleConsumerGroupHandler{}

		err := group.Consume(ctx, topics, handler)
		if err != nil {
			fmt.Println("recv error:%v", err)
		}
	}
}
