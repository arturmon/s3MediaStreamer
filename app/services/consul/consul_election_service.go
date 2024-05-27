package consul

import (
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"time"

	election "github.com/arturmon/consul-leader-election"
	"github.com/hashicorp/consul/api"
)

const checkConsulLeaderTimeoutSeconds = 5

type ConsulElection interface {
	ReElection(clien *election.Election) error
	IsLeader() bool
	GetElectionClient() *election.Election
	Init()
}

type Election struct {
	Notify   *Notify
	Election *election.Election
}

type Notify struct {
	T      string
	Logger *logs.Logger
}

func NewNotify(t string, logger *logs.Logger) *Notify {
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
	Name         string
	Key          string
	LogLevel     uint8
	Event        *Notify
}

func InitializeLeaderElection(config *LeaderElectionConfig) *election.Election {
	electionConfig := &election.ElectionConfig{
		CheckTimeout: config.CheckTimeout,
		Client:       config.Client,
		Checks:       config.Checks,
		Name:         config.Name,
		Key:          config.Key,
		LogLevel:     config.LogLevel,
		Event:        config.Event,
	}

	return election.NewElection(electionConfig)
}

func NewElection(appName string, logger *logs.Logger, client ConsulService) *Election {
	n := NewNotify(appName, logger)
	check := "service:" + appName + "-" + client.GetHostname() + ":1"
	key := "service/" + appName + "/leader"

	err := CreateOrUpdateLeaderKey(client.GetConsulClient(), logger, key, "")
	if err != nil {
		logger.Error("Error consul create kv leader")
	}

	electionConfig := &LeaderElectionConfig{
		CheckTimeout: checkConsulLeaderTimeoutSeconds * time.Second,
		Client:       client.GetConsulClient(),
		Checks:       []string{check},
		Name:         appName + "-" + client.GetHostname(),
		Key:          key,
		LogLevel:     election.LogDebug,
		Event:        n,
	}

	leaderElection := InitializeLeaderElection(electionConfig)

	err = ReadSessionInfoOnKey(logger, client.GetConsulClient())

	return &Election{
		Notify:   n,
		Election: leaderElection,
	}
}

func (r *Election) ReElection(clien *election.Election) error {
	err := clien.ReElection()
	if err != nil {
		return err
	}
	return nil
}

func (r *Election) IsLeader() bool {
	return r.Election.IsLeader()
}

func (r *Election) Init() {
	r.Election.Init()
}

func CreateOrUpdateLeaderKey(consulClient *api.Client, logger *logs.Logger, key, value string) error {
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

func ReadSessionInfoOnKey(logger *logs.Logger, consulClient *api.Client) error {
	listSession, _, err := consulClient.Session().List(nil)
	if err != nil {
		return err
	}
	for _, session := range listSession {
		logger.Printf("Session ID: %s Node: %s Name: %s CreateIndex: %d", session.ID, session.Node, session.Name, session.CreateIndex)
		// Print more session information if needed
	}
	return nil
}

func (r *Election) GetElectionClient() *election.Election {
	return r.Election
}
