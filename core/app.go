package core

import (
	"context"
	"errors"
	"github.com/CodFrm/learnMicroService/common"
	"github.com/Shopify/sarama"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

type HttpApi struct {
	Pattern string
	Handler func(http.ResponseWriter, *http.Request)
}

type AppConfig struct {
	Http    HttpConfig                //HTTP服务器配置
	Service map[string]common.Service //需要注册的服务配置
	Db      map[string]DbConfig       //数据库
	Mq      MqConfig                  //kafka消息队列配置
}

type MqConfig struct {
	Addrs []string
}

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Pwd      string
	Database string
}

type HttpConfig struct {
	Port int
	Api  []HttpApi
}

var appConfig AppConfig
var rpcService map[string]*grpc.ClientConn
var dbConnect map[string]*common.Db
var mqClient sarama.Client
//开启服务
func StartService(config AppConfig, success func()) error {
	appConfig = config
	rpcService = make(map[string]*grpc.ClientConn)
	dbConnect = make(map[string]*common.Db)
	for _, item := range config.Service {
		if item.Address == "" {
			continue
		}
		//err := RegisterService(item)
		//if err != nil {
		//	return err
		//}
	}
	for key, item := range config.Db {
		db, err := ConnectDb(item)
		if err != nil {
			return err
		}
		dbConnect[key] = db
	}
	if len(config.Mq.Addrs) > 0 {
		err := ConnectMq(config.Mq)
		if err != nil {
			return err
		}
	}
	//其实这里应该是有问题的,http不一定成功了
	success()
	return StartHttp(config.Http.Port, config.Http.Api)
}

func ConnectMq(mq MqConfig) error {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	client, err := sarama.NewClient(mq.Addrs, config)
	if err != nil {
		return err
	}
	mqClient = client
	return nil
}

var produce sarama.SyncProducer
//通过消息队列发送一条消息
func SendMessage(topic string, value string) error {
	if produce == nil {
		var err error
		produce, err = sarama.NewSyncProducerFromClient(mqClient)
		if err != nil {
			return err
		}
	}
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(value)}
	_, _, err := produce.SendMessage(msg)
	return err
}

type consumerGroupHandler struct {
	callback func(msg *sarama.ConsumerMessage) bool
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if h.callback(msg) {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}

//生成一个消费群组
func RecvMessage(groupId string,
	topic []string,
	callback func(msg *sarama.ConsumerMessage) bool) error {
	consumer, err := sarama.NewConsumerGroupFromClient(groupId, mqClient)
	if err != nil {
		return err
	}
	go func() {
		ctx := context.Background()
		handler := consumerGroupHandler{callback: callback}
		for {
			err := consumer.Consume(ctx, topic, handler)
			if err != nil {
				break
			}
		}
	}()
	return nil
}

func StartHttp(port int, api []HttpApi) error {
	for _, item := range api {
		http.HandleFunc(item.Pattern, item.Handler)
	}
	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func RegisterService(service common.Service) error {
	defer service.Deregister()
	return service.Register()
}

func ConnectDb(config DbConfig) (*common.Db, error) {
	db := &common.Db{}
	err := db.Connect(config.Host, config.Port, config.User, config.Pwd, config.Database)
	return db, err
}

//获取RPC服务客户链接
func GetRPCService(serviceName string) (*grpc.ClientConn, error) {
	if v, ok := rpcService[serviceName]; ok {
		return v, nil
	}
	service, ok := appConfig.Service[serviceName]
	if !ok {
		return nil, errors.New("not find service")
	}
	rpcConn, err := service.GetRPCService() //直接返回rpc
	if err != nil {
		return nil, err
	}
	rpcService[serviceName] = rpcConn
	return rpcConn, nil
}

func GetDb(name string) (*common.Db, error) {
	if v, ok := dbConnect[name]; ok {
		return v, nil
	}
	return nil, errors.New("not find db")
}
