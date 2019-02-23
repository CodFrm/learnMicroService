package core

import (
	"errors"
	"github.com/CodFrm/learnMicroService/common"
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
	Db      map[string]DbConfig
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

//开启服务
func StartService(config AppConfig, success func()) error {
	appConfig = config
	rpcService = make(map[string]*grpc.ClientConn)
	dbConnect = make(map[string]*common.Db)
	for _, item := range config.Service {
		if item.Address == "" {
			continue
		}
		err := RegisterService(item)
		if err != nil {
			return err
		}
	}
	for key, item := range config.Db {
		db, err := ConnectDb(item)
		if err != nil {
			return err
		}
		dbConnect[key] = db
	}
	//其实这里应该是有问题的,http不一定成功了
	success()
	return StartHttp(config.Http.Port, config.Http.Api)
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
