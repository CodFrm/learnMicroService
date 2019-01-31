package common

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/hashicorp/consul/api"
)

const (
	consul_ip = "172.28.1.3"
)

type Service struct {
	Name      string
	Tags      []string
	Address   string
	Port      int
	ServiceId string
}

func GetConsulAgent() (*api.Agent, error) {
	config := api.DefaultConfig()
	config.Address = consul_ip + ":8500"
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	agent := client.Agent()
	return agent, nil
}

//注册服务
func (s *Service) Register() error {
	agent, err := GetConsulAgent()
	if err != nil {
		return err
	}
	s.ServiceId = fmt.Sprintf("%v-%v-%v", s.Name, s.Address, s.Port)
	reg := &api.AgentServiceRegistration{
		ID:      s.ServiceId,
		Name:    s.Name,
		Tags:    s.Tags,
		Port:    s.Port,
		Address: s.Address,
		Check: &api.AgentServiceCheck{
			TCP:      fmt.Sprintf("%v:%d", s.Address, s.Port),
			Interval: "10s",
			Timeout:  "1s",
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		return err
	}
	return nil
}

//注销服务
func (s *Service) Deregister() error {
	agent, err := GetConsulAgent()
	if err != nil {
		return err
	}
	agent.ServiceDeregister(s.ServiceId)
	return nil
}

//获取rpc服务(服务发现)
func (s *Service) GetRPCService() (*grpc.ClientConn, error) {
	var host string
	for _, val := range s.Tags {
		host += val + "."
	}
	host += s.Name + ".service.consul"
	conn, err := grpc.Dial("dns://"+consul_ip+":8600/"+host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, err
}

func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
