package consul

import (
	"fmt"
	"net"
	"os"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"strconv"
	"time"

	election "github.com/dmitriyGarden/consul-leader-election"
	"github.com/hashicorp/consul/api"
)

type Notify struct {
	T      string
	Logger *logging.Logger
}

func NewNotify(t string, logger *logging.Logger) *Notify {
	return &Notify{
		T:      t,
		Logger: logger,
	}
}

func (n *Notify) EventLeader(f bool) {
	if f {
		n.Logger.Info(fmt.Sprintf("%s I'm the leader!", n.T))
	} else {
		n.Logger.Info(fmt.Sprintf("%s I'm no longer the leader!", n.T))
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

// RegisterService registers the service in Consul.
func RegisterService(client *api.Client, appName string, cfg *config.Config) error {
	port, err := strconv.Atoi(cfg.Listen.Port)
	if err != nil {
		return err // handle error appropriately
	}
	ip := GetLocalIP()

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Ошибка при получении имени хоста:", err)
		return err
	}

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      appName + "-" + hostname,
		Name:    appName,
		Port:    port,
		Address: ip, // Change to your actual service address
		Tags:    []string{"microservice", "golang"},
		Checks: api.AgentServiceChecks{
			&api.AgentServiceCheck{
				HTTP:     fmt.Sprintf("http://%s:%v/health/readiness", ip, port),
				Interval: "3s",
				Timeout:  "30s",
			},
			&api.AgentServiceCheck{
				HTTP:     fmt.Sprintf("http://%s:%v/health/liveness", ip, port),
				Interval: "10s",
				Timeout:  "30s",
			},
		},
	}
	return client.Agent().ServiceRegister(serviceRegistration)
}

// DeregisterService deregisters the service from Consul.
func DeregisterService(client *api.Client, serviceID string) error {
	return client.Agent().ServiceDeregister(serviceID)
}

func ReElection(clien *election.Election) {
	err := clien.ReElection()
	if err != nil {
		return
	}
}

func CreateOrUpdateLeaderKey(consulClient *api.Client, logger *logging.Logger, key, value string) error {
	// Check if the leader key already exists.
	existingPair, _, err := consulClient.KV().Get(key, nil)
	if err != nil {
		logger.Errorf("Failed to check if leader key exists in Consul: %v", err)
		return err
	}

	if existingPair == nil {
		// Leader key does not exist, create it
		kvPair := &api.KVPair{
			Key:   key,
			Value: []byte(value),
		}

		_, err = consulClient.KV().Put(kvPair, nil)
		if err != nil {
			logger.Errorf("Failed to create leader key in Consul: %v", err)
			return err
		}
	} else {
		// Leader key already exists, handle accordingly (perhaps log a message or take other actions).
		logger.Infof("Leader key '%s' already exists in Consul", key)
	}

	return nil
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it.
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
