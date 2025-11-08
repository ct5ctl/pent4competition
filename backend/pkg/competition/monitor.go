package competition

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/database"

	"github.com/sirupsen/logrus"
)

// FlagPattern matches FLAG{...} or flag{...}
var FlagPattern = regexp.MustCompile(`(?i)flag\{[^}]+\}`)

// FlowMonitor monitors a Flow for FLAG detection
type FlowMonitor struct {
	flowID        int64
	challengeCode string
	db            database.Querier
	client        *Client
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *logrus.Entry
	resultDir     string
	stopChan      chan struct{}
	wg            sync.WaitGroup
	lastCheckID   int64
	foundFlag     string
	flagMutex     sync.RWMutex
}

// NewFlowMonitor creates a new flow monitor
func NewFlowMonitor(
	ctx context.Context,
	flowID int64,
	challengeCode string,
	db database.Querier,
	client *Client,
	resultDir string,
) *FlowMonitor {
	monitorCtx, cancel := context.WithCancel(ctx)
	
	return &FlowMonitor{
		flowID:        flowID,
		challengeCode: challengeCode,
		db:            db,
		client:        client,
		ctx:           monitorCtx,
		cancel:        cancel,
		logger: logrus.WithFields(logrus.Fields{
			"component":      "flow-monitor",
			"flow_id":        flowID,
			"challenge_code": challengeCode,
		}),
		resultDir: resultDir,
		stopChan:  make(chan struct{}),
	}
}

// Start begins monitoring the flow
func (fm *FlowMonitor) Start() {
	fm.wg.Add(1)
	go fm.monitor()
}

// Stop stops the monitor
func (fm *FlowMonitor) Stop() {
	fm.cancel()
	close(fm.stopChan)
	fm.wg.Wait()
}

// GetFoundFlag returns the found flag if any
func (fm *FlowMonitor) GetFoundFlag() string {
	fm.flagMutex.RLock()
	defer fm.flagMutex.RUnlock()
	return fm.foundFlag
}

// monitor continuously checks for new assistant logs
func (fm *FlowMonitor) monitor() {
	defer fm.wg.Done()
	
	ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds
	defer ticker.Stop()
	
	fm.logger.Info("flow monitor started")
	
	for {
		select {
		case <-fm.ctx.Done():
			fm.logger.Info("flow monitor stopped by context")
			return
		case <-fm.stopChan:
			fm.logger.Info("flow monitor stopped")
			return
		case <-ticker.C:
			if err := fm.checkForFlags(); err != nil {
				fm.logger.WithError(err).Error("failed to check for flags")
			}
		}
	}
}

// checkForFlags checks assistant logs for FLAG patterns
func (fm *FlowMonitor) checkForFlags() error {
	// Get all assistants for this flow
	assistants, err := fm.db.GetFlowAssistants(fm.ctx, fm.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow assistants: %w", err)
	}
	
	// Check logs from all assistants
	for _, assistant := range assistants {
		logs, err := fm.db.GetFlowAssistantLogs(fm.ctx, database.GetFlowAssistantLogsParams{
			FlowID:      fm.flowID,
			AssistantID: assistant.ID,
		})
		if err != nil {
			fm.logger.WithError(err).Warn("failed to get assistant logs")
			continue
		}
		
		// Check logs newer than last checked
		for _, log := range logs {
			if log.ID <= fm.lastCheckID {
				continue
			}
			
			// Check message content for FLAG
			flags := fm.extractFlags(log.Message)
			if len(flags) > 0 {
				fm.logger.WithFields(logrus.Fields{
					"log_id": log.ID,
					"flags":  flags,
				}).Info("found FLAG in assistant log")
				
				// Try each flag
				for _, flag := range flags {
					if fm.trySubmitFlag(flag, log.Message) {
						fm.lastCheckID = log.ID
						return nil // Stop checking after successful submission
					}
				}
			}
			
			fm.lastCheckID = log.ID
		}
	}
	
	return nil
}

// extractFlags extracts all FLAG patterns from text
func (fm *FlowMonitor) extractFlags(text string) []string {
	matches := FlagPattern.FindAllString(text, -1)
	
	// Deduplicate and normalize
	flagMap := make(map[string]bool)
	var flags []string
	
	for _, match := range matches {
		// Normalize to FLAG{...} format
		normalized := strings.ToUpper(match[:4]) + match[4:]
		if !flagMap[normalized] {
			flagMap[normalized] = true
			flags = append(flags, normalized)
		}
	}
	
	return flags
}

// trySubmitFlag attempts to submit a flag and save results
func (fm *FlowMonitor) trySubmitFlag(flag, context string) bool {
	fm.logger.WithField("flag", flag).Info("attempting to submit flag")
	
	// Submit to competition API
	resp, err := fm.client.SubmitAnswer(fm.ctx, fm.challengeCode, flag)
	if err != nil {
		fm.logger.WithError(err).Error("failed to submit flag")
		return false
	}
	
	fm.logger.WithFields(logrus.Fields{
		"correct":       resp.Correct,
		"earned_points": resp.EarnedPoints,
		"is_solved":     resp.IsSolved,
	}).Info("flag submission result")
	
	// Save result to file
	if err := fm.saveResult(flag, resp, context); err != nil {
		fm.logger.WithError(err).Error("failed to save result")
	}
	
	// If correct, mark as found and stop
	if resp.Correct {
		fm.flagMutex.Lock()
		fm.foundFlag = flag
		fm.flagMutex.Unlock()
		
		fm.logger.WithFields(logrus.Fields{
			"flag":          flag,
			"earned_points": resp.EarnedPoints,
		}).Info("successfully found and submitted FLAG!")
		
		return true
	}
	
	return false
}

// ResultData represents the data saved to file
type ResultData struct {
	Timestamp     time.Time `json:"timestamp"`
	ChallengeCode string    `json:"challenge_code"`
	FlowID        int64     `json:"flow_id"`
	Flag          string    `json:"flag"`
	Correct       bool      `json:"correct"`
	EarnedPoints  int       `json:"earned_points"`
	IsSolved      bool      `json:"is_solved"`
	Context       string    `json:"context"`
}

// saveResult saves the submission result to a local file
func (fm *FlowMonitor) saveResult(flag string, resp *SubmitAnswerResponse, context string) error {
	// Create result directory if not exists
	if err := os.MkdirAll(fm.resultDir, 0755); err != nil {
		return fmt.Errorf("failed to create result directory: %w", err)
	}
	
	// Prepare result data
	result := ResultData{
		Timestamp:     time.Now(),
		ChallengeCode: fm.challengeCode,
		FlowID:        fm.flowID,
		Flag:          flag,
		Correct:       resp.Correct,
		EarnedPoints:  resp.EarnedPoints,
		IsSolved:      resp.IsSolved,
		Context:       context,
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%d.json", timestamp, fm.challengeCode, fm.flowID)
	filepath := filepath.Join(fm.resultDir, filename)
	
	// Marshal to JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}
	
	fm.logger.WithField("file", filepath).Info("saved result to file")
	
	return nil
}

