package consul_service

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
	"net"
	"s3MediaStreamer/app/internal/config"
	logging "s3MediaStreamer/app/pkg/logging"
	"strconv"
	"time"
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
	logger       *logging.Logger
	cfg          *config.Config
	AppName      string
}

func NewService(appName string, cfg *config.Config, logger *logging.Logger) ConsulService {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Consul.URL
	consulConfig.WaitTime = time.Duration(cfg.Consul.WaitTime) * time.Second
	client, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Fatal(err)
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
		s.logger.Fatal(err) // handle error appropriately
	}
	ip := s.GetLocalIP()

	checks := []*api.AgentServiceCheck{
		{
			Name:     "Readiness",
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health/readiness",
			Interval: "3s",
			Timeout:  "30s",
			Notes:    "readiness probe",
		},
		{
			Name:     "Liveness",
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health/liveness",
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
		ID:      s.AppName + "-" + s.GetHostname(),
		Name:    s.AppName,
		Tags:    []string{"microservice", "golang"},
		Address: ip,
		Port:    port,
		Checks:  checks,
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