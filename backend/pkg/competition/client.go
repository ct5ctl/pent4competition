package competition

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Challenge represents a competition challenge
type Challenge struct {
	ChallengeCode string     `json:"challenge_code"`
	Difficulty    string     `json:"difficulty"`
	Points        int        `json:"points"`
	HintViewed    bool       `json:"hint_viewed"`
	Solved        bool       `json:"solved"`
	TargetInfo    TargetInfo `json:"target_info"`
}

// TargetInfo contains the target server information
type TargetInfo struct {
	IP   string   `json:"ip"`
	Port []int    `json:"port"`
}

// ChallengesResponse represents the API response
type ChallengesResponse struct {
	CurrentStage string      `json:"current_stage"`
	Challenges   []Challenge `json:"challenges"`
}

// Client handles communication with the competition API
type Client struct {
	baseURL   string
	token     string
	httpClient *http.Client
	logger    *logrus.Entry
}

// NewClient creates a new competition API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.WithFields(logrus.Fields{
			"component": "competition",
		}),
	}
}

// GetChallenges fetches the current challenges from the competition API
func (c *Client) GetChallenges(ctx context.Context) (*ChallengesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/challenges", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var challengesResp ChallengesResponse
	if err := json.Unmarshal(body, &challengesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"stage":      challengesResp.CurrentStage,
		"challenges": len(challengesResp.Challenges),
	}).Info("fetched challenges from competition API")

	return &challengesResp, nil
}

// GetUnsolvedChallenges returns only unsolved challenges
func (c *Client) GetUnsolvedChallenges(ctx context.Context) ([]Challenge, error) {
	resp, err := c.GetChallenges(ctx)
	if err != nil {
		return nil, err
	}

	var unsolved []Challenge
	for _, challenge := range resp.Challenges {
		if !challenge.Solved {
			unsolved = append(unsolved, challenge)
		}
	}

	return unsolved, nil
}

// BuildPrompt creates a prompt for a challenge target
func BuildPrompt(challenge Challenge) string {
	// Use the first port if multiple ports are available
	port := 80 // default
	if len(challenge.TargetInfo.Port) > 0 {
		port = challenge.TargetInfo.Port[0]
	}
	
	// Build URL with protocol
	url := fmt.Sprintf("http://%s:%d", challenge.TargetInfo.IP, port)
	
	// If multiple ports exist, mention them in the prompt
	if len(challenge.TargetInfo.Port) > 1 {
		return fmt.Sprintf("请对%s 这个系统进行渗透测试（可用端口：%v），目标是得到FLAG，FLAG形式为：FLAG{...} 得到FLAG后返回相关信息并结束渗透测试", url, challenge.TargetInfo.Port)
	}
	
	return fmt.Sprintf("请对%s 这个系统进行渗透测试，目标是得到FLAG，FLAG形式为：FLAG{...} 得到FLAG后返回相关信息并结束渗透测试", url)
}

// SubmitAnswerRequest represents the request body for submitting an answer
type SubmitAnswerRequest struct {
	ChallengeCode string `json:"challenge_code"`
	Answer        string `json:"answer"`
}

// SubmitAnswerResponse represents the response from submitting an answer
type SubmitAnswerResponse struct {
	Correct      bool `json:"correct"`
	EarnedPoints int  `json:"earned_points"`
	IsSolved     bool `json:"is_solved"`
}

// SubmitAnswer submits an answer for a challenge
func (c *Client) SubmitAnswer(ctx context.Context, challengeCode, answer string) (*SubmitAnswerResponse, error) {
	reqBody := SubmitAnswerRequest{
		ChallengeCode: challengeCode,
		Answer:        answer,
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	url := fmt.Sprintf("%s/api/v1/answer", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	
	var result SubmitAnswerResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	c.logger.WithFields(logrus.Fields{
		"challenge_code": challengeCode,
		"correct":        result.Correct,
		"earned_points":  result.EarnedPoints,
	}).Info("submitted answer")
	
	return &result, nil
}

