package consul

import (
	"net"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
)

const (
	ttl          = time.Second * 8
	healthTicket = time.Second * 5
)

type ConsulService interface {
	Start()
	UpdateHealthCheck()
	RegisterService()
	SetupConsulWatch()
	GetLocalIP() string
	GetHostname() string
	GetConsulClient() *api.Client
}

type Service struct {
	ConsulClient *api.Client
	logger       *logs.Logger
	cfg          *model.Config
	AppName      string
}

func NewService(appName string, cfg *model.Config, logger *logs.Logger) ConsulService {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Consul.URL
	consulConfig.WaitTime = time.Duration(cfg.Consul.WaitTime) * time.Second
	client, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Fatal("Failed to create Consul client:", err)
	}

	for {
		_, err = client.Agent().Services()
		if err != nil {
			logger.Errorf("Failed to fetch Consul services: %s", err)
			logger.Infof("Retrying connection to Consul in %d seconds...", cfg.MessageQueue.RetryingConnection)
			time.Sleep(time.Duration(cfg.MessageQueue.RetryingConnection) * time.Second)
			continue
		}
		break
	}

	return &Service{
		ConsulClient: client,
		logger:       logger,
		cfg:          cfg,
		AppName:      appName,
	}
}

func (s *Service) Start() {
	s.RegisterService()
	go s.UpdateHealthCheck()
	s.SetupConsulWatch()
}

func (s *Service) UpdateHealthCheck() {
	ticker := time.NewTicker(healthTicket)
	check := "service:" + s.AppName + "-" + s.GetHostname() + ":3"
	for {
		err := s.ConsulClient.Agent().UpdateTTL(check, "online", api.HealthPassing)
		if err != nil {
			s.logger.Fatal(err)
		}
		<-ticker.C
	}
}

func (s *Service) RegisterService() {
	port, err := strconv.Atoi(s.cfg.Listen.Port)
	if err != nil {
		s.logger.Fatal("Failed to register Consul services:", err) // handle error appropriately
	}
	ip := s.GetLocalIP()

	connect := &api.AgentServiceConnect{
		Native: true,
	}

	checks := []*api.AgentServiceCheck{
		{
			Name:     "Readiness",
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health_handler/readiness",
			Interval: "3s",
			Timeout:  "30s",
			Notes:    "readiness probe",
		},
		{
			Name:     "Liveness",
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health_handler/liveness",
			Interval: "10s",
			Timeout:  "30s",
			Notes:    "liveness probe",
		},
		{
			Name:                           "TTL probe",
			DeregisterCriticalServiceAfter: ttl.String(),
			TLSSkipVerify:                  true,
			TTL:                            ttl.String(),
			Notes:                          "TTL probe",
		},
	}

	register := &api.AgentServiceRegistration{
		Meta:    s.setMetadata(),
		ID:      s.AppName + "-" + s.GetHostname(),
		Name:    s.AppName,
		Tags:    []string{"microservice", "golang"},
		Address: ip,
		Port:    port,
		Checks:  checks,
		Connect: connect,
	}

	err = s.ConsulClient.Agent().ServiceRegister(register)
	if err != nil {
		s.logger.Fatal(err)
	}
}

func (s *Service) SetupConsulWatch() {
	query := map[string]interface{}{
		"type":        "service",
		"service":     s.AppName,
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		s.logger.Fatal(err)
	}

	plan.HybridHandler = func(index watch.BlockingParamVal, result interface{}) {
		if msg, ok := result.([]*api.ServiceEntry); ok {
			for _, entry := range msg {
				s.logger.Debugln("new member joined", entry.Service)
			}
		}
	}

	var watchLogger hclog.Logger
	go func() {
		err = plan.RunWithClientAndHclog(s.ConsulClient, watchLogger)
		if err != nil {
			return
		}
	}()
}

func (s *Service) GetConsulClient() *api.Client {
	return s.ConsulClient
}

func (s *Service) setMetadata() map[string]string {
	s3 := s.cfg.AppConfig.S3.Endpoint + "/" + s.cfg.AppConfig.S3.BucketName
	caching := strconv.FormatBool(s.cfg.Storage.Caching.Enabled) +
		", " + s.cfg.Storage.Caching.Address
	openTelemetry := strconv.FormatBool(s.cfg.AppConfig.OpenTelemetry.TracingEnabled) +
		", " + s.cfg.AppConfig.OpenTelemetry.JaegerEndpoint
	storage := s.cfg.Storage.Host +
		":" + s.cfg.Storage.Port +
		"/" + s.cfg.Storage.Database
	var sessionStorage string
	switch s.cfg.Session.SessionStorageType {
	case "postgres":
		sessionStorage = s.cfg.Session.Postgresql.PostgresqlHost +
			":" + s.cfg.Session.Postgresql.PostgresqlPort +
			"/" + s.cfg.Session.Postgresql.PostgresqlDatabase
	case "mongodb":
		sessionStorage = s.cfg.Session.Mongodb.MongoHost +
			":" + s.cfg.Session.Mongodb.MongoPort +
			"/" + s.cfg.Session.Mongodb.MongoDatabase
	case "cookies":
		sessionStorage = "-"
	case "memcached":
		sessionStorage = s.cfg.Session.Memcached.MemcachedHost +
			":" + s.cfg.Session.Memcached.MemcachedPort
	}
	return map[string]string{
		"type":            "api",
		"log-level":       s.cfg.AppConfig.GinMode,
		"s3":              s3,
		"caching":         caching,
		"session-type":    s.cfg.Session.SessionStorageType,
		"openTelemetry":   openTelemetry,
		"storage":         storage,
		"session-Storage": sessionStorage,
	}
}
