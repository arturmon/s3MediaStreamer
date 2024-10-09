package consul

import (
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/hashicorp/consul/api"
)

type InterfaceKV interface {
	GetFromConsul(key string) ([]byte, error)
	PutToConsul(key string, value string) error
	FetchConsulConfig(key string, defaultValue string) (string, error)
}

type KVService struct {
	ConsulClient *api.Client
	cfg          *model.Config
	logger       *logs.Logger
}

func NewConsulKVService(cfg *model.Config, logger *logs.Logger, client Service) *KVService {

	return &KVService{
		ConsulClient: client.GetConsulClient(),
		logger:       logger,
		cfg:          cfg,
	}
}

// GetFromConsul retrieves the value for a given key from Consul.
func (k *KVService) GetFromConsul(key string) ([]byte, error) {
	kv, _, err := k.ConsulClient.KV().Get(key, nil)
	if err != nil {
		k.logger.Error("Error fetching key from Consul:", err)
		return nil, err
	}

	// Handle the case where the key is not found (kv is nil).
	if kv == nil {
		k.logger.Warnf("Key not found in Consul: %s", key)
		return nil, nil // Return nil, indicating the key does not exist.
	}

	return kv.Value, nil
}

// PutToConsul sets a value in Consul for a given key.
func (k *KVService) PutToConsul(key string, value string) error {
	kvPair := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}
	_, err := k.ConsulClient.KV().Put(kvPair, nil)
	if err != nil {
		k.logger.Error("Error setting key in Consul:", err)
		return err
	}
	return nil
}

// FetchConsulConfig fetches the job interval configuration from Consul.
func (k *KVService) FetchConsulConfig(key string, defaultValue string) (string, error) {
	value, err := k.GetFromConsul(key)
	if err != nil {
		return "", err
	}

	if value == nil {
		k.logger.Warnf("Key not found in Consul, creating with default value: %s", key)
		err = k.PutToConsul(key, defaultValue)
		if err != nil {
			return "", err
		}
		return defaultValue, nil
	}

	return string(value), nil
}
