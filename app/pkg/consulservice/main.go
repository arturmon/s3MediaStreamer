package consulservice

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
	"net"
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
	s.setupConsulWatch()
}

func (s *Service) updateHealthCheck() {
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

func (s *Service) registerService() {
	port, err := strconv.Atoi(s.cfg.Listen.Port)
	if err != nil {
		s.logger.Fatal(err) // handle error appropriately
	}
	ip := s.GetLocalIP()

	checks := []*api.AgentServiceCheck{
		{
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health/readiness",
			Interval: "3s",
			Timeout:  "30s",
			Notes:    "readiness probe",
		},
		{
			HTTP:     "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health/liveness",
			Interval: "10s",
			Timeout:  "30s",
			Notes:    "liveness probe",
		},
		{
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

func (s *Service) setupConsulWatch() {
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
		switch msg := result.(type) {
		case []*api.ServiceEntry:
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
