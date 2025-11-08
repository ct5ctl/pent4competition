package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Challenge represents a challenge
type Challenge struct {
	ChallengeCode string     `json:"challenge_code"`
	Difficulty    string     `json:"difficulty"`
	Points        int        `json:"points"`
	HintViewed    bool       `json:"hint_viewed"`
	Solved        bool       `json:"solved"`
	TargetInfo    TargetInfo `json:"target_info"`
}

// TargetInfo contains target information
type TargetInfo struct {
	IP   string `json:"ip"`
	Port []int  `json:"port"`
}

// ChallengesResponse is the response for challenges endpoint
type ChallengesResponse struct {
	CurrentStage string      `json:"current_stage"`
	Challenges   []Challenge `json:"challenges"`
}

// SubmitAnswerRequest is the request for submitting an answer
type SubmitAnswerRequest struct {
	ChallengeCode string `json:"challenge_code"`
	Answer        string `json:"answer"`
}

// SubmitAnswerResponse is the response for submitting an answer
type SubmitAnswerResponse struct {
	Correct      bool `json:"correct"`
	EarnedPoints int  `json:"earned_points"`
	IsSolved     bool `json:"is_solved"`
}

// FlagSubmission records a flag submission
type FlagSubmission struct {
	Timestamp     time.Time `json:"timestamp"`
	ChallengeCode string    `json:"challenge_code"`
	Answer        string    `json:"answer"`
	Correct       bool      `json:"correct"`
	EarnedPoints  int       `json:"earned_points"`
}

// MockAPI is the mock competition API server
type MockAPI struct {
	submissions   []FlagSubmission
	submissionsMu sync.RWMutex
	correctFlags  map[string]string // challenge_code -> correct flag
	solvedFlags   map[string]bool   // track solved challenges
	outputDir     string
}

// NewMockAPI creates a new mock API server
func NewMockAPI(correctFlags map[string]string, outputDir string) *MockAPI {
	return &MockAPI{
		submissions:  make([]FlagSubmission, 0),
		correctFlags: correctFlags,
		solvedFlags:  make(map[string]bool),
		outputDir:    outputDir,
	}
}

// GetChallenges handles GET /api/v1/challenges
func (m *MockAPI) GetChallenges(c *gin.Context) {
	// Return a mock challenge
	response := ChallengesResponse{
		CurrentStage: "debug",
		Challenges: []Challenge{
			{
				ChallengeCode: "debug_test",
				Difficulty:    "debug",
				Points:        100,
				HintViewed:    false,
				Solved:        m.solvedFlags["debug_test"],
				TargetInfo: TargetInfo{
					IP:   "127.0.0.1",
					Port: []int{8080},
				},
			},
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// SubmitAnswer handles POST /api/v1/answer
func (m *MockAPI) SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数校验失败"})
		return
	}
	
	// Check if the answer is correct
	correctFlag, exists := m.correctFlags[req.ChallengeCode]
	isCorrect := exists && req.Answer == correctFlag
	
	// Calculate points
	earnedPoints := 0
	if isCorrect {
		earnedPoints = 100 // Default points
	}
	
	// Track if already solved
	isSolved := m.solvedFlags[req.ChallengeCode]
	if isCorrect {
		m.solvedFlags[req.ChallengeCode] = true
	}
	
	// Record submission
	submission := FlagSubmission{
		Timestamp:     time.Now(),
		ChallengeCode: req.ChallengeCode,
		Answer:        req.Answer,
		Correct:       isCorrect,
		EarnedPoints:  earnedPoints,
	}
	
	m.submissionsMu.Lock()
	m.submissions = append(m.submissions, submission)
	m.submissionsMu.Unlock()
	
	// Save to file
	if err := m.saveSubmission(submission); err != nil {
		log.Printf("Failed to save submission: %v", err)
	}
	
	// Log to console
	status := "❌ INCORRECT"
	if isCorrect {
		status = "✅ CORRECT"
	}
	
	log.Printf("=== FLAG SUBMISSION ===")
	log.Printf("Challenge: %s", req.ChallengeCode)
	log.Printf("Answer:    %s", req.Answer)
	log.Printf("Status:    %s", status)
	log.Printf("Points:    %d", earnedPoints)
	log.Printf("=====================")
	
	// Return response
	response := SubmitAnswerResponse{
		Correct:      isCorrect,
		EarnedPoints: earnedPoints,
		IsSolved:     isSolved,
	}
	
	c.JSON(http.StatusOK, response)
}

// saveSubmission saves a submission to file
func (m *MockAPI) saveSubmission(submission FlagSubmission) error {
	// Create output directory if not exists
	if err := os.MkdirAll(m.outputDir, 0755); err != nil {
		return err
	}
	
	// Generate filename
	timestamp := submission.Timestamp.Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", timestamp, submission.ChallengeCode)
	filepath := filepath.Join(m.outputDir, filename)
	
	// Marshal to JSON
	data, err := json.MarshalIndent(submission, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(filepath, data, 0644)
}

// GetSubmissions handles GET /submissions (additional endpoint for viewing)
func (m *MockAPI) GetSubmissions(c *gin.Context) {
	m.submissionsMu.RLock()
	defer m.submissionsMu.RUnlock()
	
	c.JSON(http.StatusOK, gin.H{
		"total":       len(m.submissions),
		"submissions": m.submissions,
	})
}

func main() {
	// Configuration
	port := os.Getenv("MOCK_API_PORT")
	if port == "" {
		port = "8000"
	}
	
	outputDir := os.Getenv("MOCK_API_OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "./mock_api_submissions"
	}
	
	// Define correct flags for testing
	correctFlags := map[string]string{
		"debug_test": "FLAG{test_flag_12345}",
		// Add more test flags as needed
	}
	
	// Create mock API
	mockAPI := NewMockAPI(correctFlags, outputDir)
	
	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})
	
	// Routes
	api := r.Group("/api/v1")
	{
		api.GET("/challenges", mockAPI.GetChallenges)
		api.POST("/answer", mockAPI.SubmitAnswer)
	}
	
	// Additional debug endpoints
	r.GET("/submissions", mockAPI.GetSubmissions)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Print startup info
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🚀 Mock Competition API Server")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Server:        http://localhost:%s\n", port)
	fmt.Printf("Output Dir:    %s\n", outputDir)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /api/v1/challenges  - Get challenges")
	fmt.Println("  POST /api/v1/answer      - Submit answer")
	fmt.Println("  GET  /submissions        - View all submissions")
	fmt.Println("  GET  /health             - Health check")
	fmt.Println("\nCorrect Flags:")
	for code, flag := range correctFlags {
		fmt.Printf("  %s: %s\n", code, flag)
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n⏳ Waiting for FLAG submissions...\n")
	
	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

