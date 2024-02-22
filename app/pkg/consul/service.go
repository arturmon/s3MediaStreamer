package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"strconv"
	"time"
)

const (
	ttl          = time.Second * 8
	healthTicket = time.Second * 5
)

type Service struct {
	ConsulClient *api.Client
	logger       *logging.Logger
	cfg          *config.Config
	AppName      string
}

func NewService(appName string, cfg *config.Config, logger *logging.Logger) *Service {
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
	s.registerService()
	go s.updateHealthCheck()
}

func (s *Service) updateHealthCheck() {
	ticker := time.NewTicker(healthTicket)
	check := "service:" + s.AppName + "-" + GetHostname() + ":3"
	for {
		err := s.ConsulClient.Agent().UpdateTTL(check, "online", api.HealthPassing)
		if err != nil {
			s.logger.Fatal(err)
		}
		<-ticker.C
	}
}

func (s *Service) registerService() {
	port, err := strconv.Atoi(s.cfg.Listen.Port)
	if err != nil {
		s.logger.Fatal(err) // handle error appropriately
	}
	ip := GetLocalIP()

	check := []*api.AgentServiceCheck{
		{
			HTTP:     fmt.Sprintf("http://%s:%v/health/readiness", ip, port),
			Interval: "3s",
			Timeout:  "30s",
		},
		{
			HTTP:     fmt.Sprintf("http://%s:%v/health/liveness", ip, port),
			Interval: "10s",
			Timeout:  "30s",
		},
		{
			DeregisterCriticalServiceAfter: ttl.String(),
			TLSSkipVerify:                  true,
			TTL:                            ttl.String(),
			//CheckID:                        checkID,
		},
	}

	register := &api.AgentServiceRegistration{
		ID:      s.AppName + "-" + GetHostname(),
		Name:    s.AppName,
		Tags:    []string{"microservice", "golang"},
		Address: ip,
		Port:    port,
		Checks:  check,
	}

	query := map[string]any{
		"type":        "service",
		"service":     s.AppName,
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		s.logger.Fatal(err)
	}

	plan.HybridHandler = func(index watch.BlockingParamVal, result any) {
		switch msg := result.(type) {
		case []*api.ServiceEntry:
			for _, entry := range msg {
				s.logger.Debugln("new member joined", entry.Service)
			}
		}

	}

	go func() {
		plan.RunWithConfig("", &api.Config{})
	}()

	err = s.ConsulClient.Agent().ServiceRegister(register)
	if err != nil {
		s.logger.Fatal(err)
	}

}
