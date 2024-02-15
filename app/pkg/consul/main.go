package consul

import (
	"fmt"
	election "github.com/dmitriyGarden/consul-leader-election"
	"github.com/hashicorp/consul/api"
	"net"
	"skeleton-golange-application/app/internal/config"
	"strconv"
	"time"
)

type Notify struct {
	T string
}

func (n *Notify) EventLeader(f bool) {
	if f {
		fmt.Println(n.T, "I'm the leader!")
	} else {
		fmt.Println(n.T, "I'm no longer the leader!")
	}
}

type LeaderElectionConfig struct {
	CheckTimeout time.Duration
	Client       *api.Client
	Checks       []string
	Key          string
	LogLevel     uint8
	Event        *Notify
}

func InitializeLeaderElection(config *LeaderElectionConfig) *election.Election {
	electionConfig := &election.ElectionConfig{
		CheckTimeout: config.CheckTimeout,
		Client:       config.Client,
		Checks:       config.Checks,
		Key:          config.Key,
		LogLevel:     config.LogLevel,
		Event:        config.Event,
	}

	return election.NewElection(electionConfig)
}

// RegisterService registers the service in Consul
func RegisterService(client *api.Client, appName string, cfg *config.Config) error {
	port, err := strconv.Atoi(cfg.Listen.Port)
	if err != nil {
		return err // handle error appropriately
	}
	ip := GetLocalIP()

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      appName,
		Name:    appName,
		Port:    port,
		Address: ip, // Change to your actual service address
		Tags:    []string{"microservice", "golang"},
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%v/health", ip, port), // Change to your actual health check endpoint
			Interval: "10s",
			Timeout:  "30s",
		},
	}
	return client.Agent().ServiceRegister(serviceRegistration)
}

// DeregisterService deregisters the service from Consul
func DeregisterService(client *api.Client, serviceID string) error {
	return client.Agent().ServiceDeregister(serviceID)
}

func ReElection(clien *election.Election) {
	for {
		time.Sleep(10 * time.Second)
		err := clien.ReElection()
		if err != nil {
			return
		}
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
