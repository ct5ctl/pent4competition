package competition

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
)

// Service manages automatic challenge processing
type Service struct {
	cfg            *config.Config
	client         *Client
	controller     controller.FlowController
	providers      providers.ProviderController
	db             database.Querier
	logger         *logrus.Entry
	processed      map[string]bool      // Track processed challenges by challenge_code
	monitors       map[int64]*FlowMonitor // Track flow monitors by flow ID
	processedMutex sync.RWMutex
	monitorsMutex  sync.RWMutex
	running        bool
	stopChan       chan struct{}
}

// NewService creates a new competition service
func NewService(
	cfg *config.Config,
	controller controller.FlowController,
	providers providers.ProviderController,
	db database.Querier,
) *Service {
	client := NewClient(cfg.CompetitionBaseURL, cfg.CompetitionToken)
	
	return &Service{
		cfg:        cfg,
		client:     client,
		controller: controller,
		providers:  providers,
		db:         db,
		logger: logrus.WithFields(logrus.Fields{
			"component": "competition-service",
		}),
		processed: make(map[string]bool),
		monitors:  make(map[int64]*FlowMonitor),
		stopChan:  make(chan struct{}),
	}
}

// Start begins the competition service
func (s *Service) Start(ctx context.Context) error {
	if !s.cfg.CompetitionEnabled {
		s.logger.Info("competition service is disabled")
		return nil
	}

	if s.cfg.CompetitionBaseURL == "" || s.cfg.CompetitionToken == "" {
		return fmt.Errorf("competition API base URL and token must be configured")
	}

	s.running = true
	s.logger.Info("starting competition service")

	go s.run(ctx)

	return nil
}

// Stop stops the competition service
func (s *Service) Stop() {
	if !s.running {
		return
	}

	s.logger.Info("stopping competition service")
	s.running = false
	close(s.stopChan)
	
	// Stop all monitors
	s.monitorsMutex.Lock()
	defer s.monitorsMutex.Unlock()
	for _, monitor := range s.monitors {
		monitor.Stop()
	}
}

// run is the main loop that periodically checks for new challenges
func (s *Service) run(ctx context.Context) {
	interval := time.Duration(s.cfg.CompetitionInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Process immediately on start
	s.processChallenges(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("competition service context cancelled")
			return
		case <-s.stopChan:
			s.logger.Info("competition service stopped")
			return
		case <-ticker.C:
			s.processChallenges(ctx)
		}
	}
}

// processChallenges fetches and processes challenges
func (s *Service) processChallenges(ctx context.Context) {
	var challenges []Challenge
	var err error
	
	// Debug mode: use configured IP and ports instead of API
	if s.cfg.CompetitionDebugMode {
		s.logger.Info("running in DEBUG mode, using configured target")
		challenges = s.getDebugChallenges()
	} else {
		// Normal mode: fetch from API
		challenges, err = s.client.GetUnsolvedChallenges(ctx)
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch challenges")
			return
		}
	}

	if len(challenges) == 0 {
		s.logger.Debug("no unsolved challenges found")
		return
	}

	s.logger.WithField("count", len(challenges)).Info("processing challenges")

	// Get default user ID (first user in the system)
	userID, err := s.getDefaultUserID(ctx)
	if err != nil {
		s.logger.WithError(err).Error("failed to get default user ID")
		return
	}

	// Get the first available provider
	prvname, prvtype, err := s.getDefaultProvider(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("failed to get default provider")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"provider_name": prvname,
		"provider_type": prvtype,
	}).Info("using provider for competition flows")

	// Process challenges sequentially
	for _, challenge := range challenges {
		// Check if already processed
		s.processedMutex.RLock()
		if s.processed[challenge.ChallengeCode] {
			s.processedMutex.RUnlock()
			s.logger.WithField("challenge_code", challenge.ChallengeCode).Debug("challenge already processed, skipping")
			continue
		}
		s.processedMutex.RUnlock()

		// Process this challenge
		if err := s.processChallenge(ctx, challenge, userID, prvname, prvtype); err != nil {
			s.logger.WithError(err).WithField("challenge_code", challenge.ChallengeCode).Error("failed to process challenge")
			continue
		}

		// Mark as processed
		s.processedMutex.Lock()
		s.processed[challenge.ChallengeCode] = true
		s.processedMutex.Unlock()

		s.logger.WithField("challenge_code", challenge.ChallengeCode).Info("challenge processing started")
	}
}

// processChallenge creates a flow for a challenge
func (s *Service) processChallenge(
	ctx context.Context,
	challenge Challenge,
	userID int64,
	prvname provider.ProviderName,
	prvtype provider.ProviderType,
) error {
	prompt := BuildPrompt(challenge)
	
	s.logger.WithFields(logrus.Fields{
		"challenge_code": challenge.ChallengeCode,
		"target_ip":      challenge.TargetInfo.IP,
		"target_ports":   challenge.TargetInfo.Port,
	}).Info("creating flow for challenge")

	// Create flow with the challenge prompt
	flow, err := s.controller.CreateFlow(ctx, userID, prompt, prvname, prvtype, nil)
	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}
	
	flowID := flow.GetFlowID()
	
	// Start monitoring this flow for FLAG detection
	resultDir := "./competition_results"
	monitor := NewFlowMonitor(ctx, flowID, challenge.ChallengeCode, s.db, s.client, resultDir)
	monitor.Start()
	
	s.monitorsMutex.Lock()
	s.monitors[flowID] = monitor
	s.monitorsMutex.Unlock()
	
	s.logger.WithFields(logrus.Fields{
		"flow_id":        flowID,
		"challenge_code": challenge.ChallengeCode,
	}).Info("flow created and monitor started")
	
	// Start a goroutine to wait for FLAG and terminate flow
	go s.monitorFlowCompletion(ctx, flowID, challenge.ChallengeCode, monitor)

	return nil
}

// monitorFlowCompletion monitors a flow and terminates it when FLAG is found
func (s *Service) monitorFlowCompletion(ctx context.Context, flowID int64, challengeCode string, monitor *FlowMonitor) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	
	timeout := time.NewTimer(30 * time.Minute) // Max 30 minutes per challenge
	defer timeout.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout.C:
			s.logger.WithFields(logrus.Fields{
				"flow_id":        flowID,
				"challenge_code": challengeCode,
			}).Warn("flow timeout reached, stopping flow")
			
			// Stop the flow
			if err := s.controller.StopFlow(ctx, flowID); err != nil {
				s.logger.WithError(err).Error("failed to stop flow")
			}
			
			// Stop monitor
			monitor.Stop()
			s.monitorsMutex.Lock()
			delete(s.monitors, flowID)
			s.monitorsMutex.Unlock()
			return
			
		case <-ticker.C:
			// Check if FLAG was found
			if flag := monitor.GetFoundFlag(); flag != "" {
				s.logger.WithFields(logrus.Fields{
					"flow_id":        flowID,
					"challenge_code": challengeCode,
					"flag":           flag,
				}).Info("FLAG found, stopping flow")
				
				// Stop the flow
				if err := s.controller.StopFlow(ctx, flowID); err != nil {
					s.logger.WithError(err).Error("failed to stop flow")
				}
				
				// Stop monitor
				monitor.Stop()
				s.monitorsMutex.Lock()
				delete(s.monitors, flowID)
				s.monitorsMutex.Unlock()
				return
			}
			
			// Check if flow is already completed/stopped
			flow, err := s.controller.GetFlow(ctx, flowID)
			if err != nil {
				s.logger.WithError(err).Error("failed to get flow status")
				continue
			}
			
			flowStatus, err := flow.GetStatus(ctx)
			if err != nil {
				s.logger.WithError(err).Error("failed to get flow status")
				continue
			}
			
			if flowStatus == database.FlowStatusFinished || flowStatus == database.FlowStatusFailed {
				s.logger.WithFields(logrus.Fields{
					"flow_id":        flowID,
					"challenge_code": challengeCode,
					"status":         flowStatus,
				}).Info("flow already finished/failed")
				
				monitor.Stop()
				s.monitorsMutex.Lock()
				delete(s.monitors, flowID)
				s.monitorsMutex.Unlock()
				return
			}
		}
	}
}

// getDefaultProvider gets the first available provider
func (s *Service) getDefaultProvider(ctx context.Context, userID int64) (provider.ProviderName, provider.ProviderType, error) {
	// Get all available providers for the user
	providers, err := s.providers.GetProviders(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get providers: %w", err)
	}

	// Try default providers in order of preference
	preferredProviders := []provider.ProviderName{
		provider.DefaultProviderNameOpenAI,
		provider.DefaultProviderNameAnthropic,
		provider.DefaultProviderNameGemini,
		provider.DefaultProviderNameBedrock,
		provider.DefaultProviderNameCustom,
		provider.DefaultProviderNameOllama,
	}

	for _, prvname := range preferredProviders {
		if prv, err := providers.Get(prvname); err == nil {
			return prvname, prv.Type(), nil
		}
	}

	// If no default provider found, try any available provider
	names := providers.ListNames()
	if len(names) > 0 {
		prv, err := providers.Get(names[0])
		if err == nil {
			return names[0], prv.Type(), nil
		}
	}

	return "", "", fmt.Errorf("no LLM provider available")
}

// getDefaultUserID gets the first user ID from the database
func (s *Service) getDefaultUserID(ctx context.Context) (int64, error) {
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return 0, fmt.Errorf("no users found in database")
	}

	// Return the first user's ID
	return users[0].ID, nil
}

// getDebugChallenges creates debug challenges from configuration
func (s *Service) getDebugChallenges() []Challenge {
	if s.cfg.CompetitionDebugIP == "" {
		s.logger.Warn("debug mode enabled but no target IP configured")
		return nil
	}
	
	// Parse ports from comma-separated string
	ports := s.parseDebugPorts(s.cfg.CompetitionDebugPorts)
	
	challenge := Challenge{
		ChallengeCode: s.cfg.CompetitionDebugCode,
		Difficulty:    "debug",
		Points:        0,
		HintViewed:    false,
		Solved:        false,
		TargetInfo: TargetInfo{
			IP:   s.cfg.CompetitionDebugIP,
			Port: ports,
		},
	}
	
	s.logger.WithFields(logrus.Fields{
		"challenge_code": challenge.ChallengeCode,
		"target_ip":      challenge.TargetInfo.IP,
		"target_ports":   challenge.TargetInfo.Port,
	}).Info("created debug challenge")
	
	return []Challenge{challenge}
}

// parseDebugPorts parses comma-separated port string to int slice
func (s *Service) parseDebugPorts(portsStr string) []int {
	if portsStr == "" {
		return []int{80}
	}
	
	parts := strings.Split(portsStr, ",")
	var ports []int
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if port, err := strconv.Atoi(part); err == nil {
			ports = append(ports, port)
		} else {
			s.logger.WithError(err).WithField("port", part).Warn("invalid port number, skipping")
		}
	}
	
	if len(ports) == 0 {
		ports = []int{80}
	}
	
	return ports
}

