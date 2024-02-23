package consul_election

import (
	"fmt"
	election "github.com/dmitriyGarden/consul-leader-election"
	"github.com/hashicorp/consul/api"
	consul_service "skeleton-golange-application/app/pkg/consul-service"
	"skeleton-golange-application/app/pkg/logging"
	"time"
)

const checkConsulLeaderTimeoutSeconds = 5

type Election struct {
	Notify   *Notify
	Election *election.Election
}

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

func NewElection(appName string, logger *logging.Logger, client *consul_service.Service) *Election {
	n := NewNotify(appName, logger)
	check := "service:" + appName + "-" + client.GetHostname() + ":1"
	key := "service/" + appName + "/leader"

	err := CreateOrUpdateLeaderKey(client.ConsulClient, logger, key, "")
	if err != nil {
		logger.Error("Error consul create kv leader")
	}

	electionConfig := &LeaderElectionConfig{
		CheckTimeout: checkConsulLeaderTimeoutSeconds * time.Second,
		Client:       client.ConsulClient,
		Checks:       []string{check},
		Key:          key,
		LogLevel:     election.LogDebug,
		Event:        n,
	}

	leaderElection := InitializeLeaderElection(electionConfig)

	return &Election{
		Notify:   n,
		Election: leaderElection,
	}
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
