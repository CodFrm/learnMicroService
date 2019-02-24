package events

import (
	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd"
	"github.com/Shopify/sarama"
)

//var event_list map[]

//注册监听事件
func RegisterEvent(event ddd.EventMessage) error {
	return core.RecvMessage(event.GetGroupId(), event.GetEventNames(), func(msg *sarama.ConsumerMessage) bool {
		err := event.Handler(msg.Topic, msg.Value)
		if err != nil {
			return false
		}
		return true
	})
}
