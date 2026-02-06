package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed static templates
var content embed.FS

// SessionsJSON represents the sessions.json file structure
type SessionsJSON map[string]SessionEntry

// SessionEntry is a single session from sessions.json
type SessionEntry struct {
	SessionID     string `json:"sessionId"`
	UpdatedAt     int64  `json:"updatedAt"`
	Label         string `json:"label,omitempty"`
	Channel       string `json:"channel"`
	Model         string `json:"model,omitempty"`
	TotalTokens   int    `json:"totalTokens"`
	ContextTokens int    `json:"contextTokens"`
	SpawnedBy     string `json:"spawnedBy,omitempty"`
}

// CronJobsFile represents the cron jobs.json structure
type CronJobsFile struct {
	Version int       `json:"version"`
	Jobs    []CronJob `json:"jobs"`
}

// CronJob represents a single cron job
type CronJob struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// TranscriptEntry represents a line in the JSONL transcript
type TranscriptEntry struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	ParentID  string          `json:"parentId"`
	Timestamp string          `json:"timestamp"`
	Message   *MessageContent `json:"message,omitempty"`
	Version   int             `json:"version,omitempty"`
	CWD       string          `json:"cwd,omitempty"`
	Provider  string          `json:"provider,omitempty"`
	ModelID   string          `json:"modelId,omitempty"`
}

// MessageContent is the actual message payload
type MessageContent struct {
	Role         string          `json:"role"`
	Content      json.RawMessage `json:"content"`
	Model        string          `json:"model,omitempty"`
	Timestamp    int64           `json:"timestamp,omitempty"`
	Usage        *UsageInfo      `json:"usage,omitempty"`
	StopReason   string          `json:"stopReason,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// UsageInfo contains token usage
type UsageInfo struct {
	Input       int       `json:"input"`
	Output      int       `json:"output"`
	CacheRead   int       `json:"cacheRead"`
	CacheWrite  int       `json:"cacheWrite"`
	TotalTokens int       `json:"totalTokens"`
	Cost        *CostInfo `json:"cost,omitempty"`
}

// CostInfo contains cost breakdown
type CostInfo struct {
	Total float64 `json:"total"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Name      string          `json:"name,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Session represents an OpenClaw session for display
type Session struct {
	SessionID    string
	Key          string
	Name         string
	Model        string
	Provider     string
	Channel      string
	MessageCount int
	TotalTokens  int
	TotalCost    float64
	TodayCost    float64
	UpdatedTime  time.Time
	IsActive     bool
	TypeBadge    string
	TypeColor    string
	Kind         string
	CWD          string
}

// DisplayMessage is a parsed message for display
type DisplayMessage struct {
	Role        string
	Content     string
	Time        time.Time
	Model       string
	Tokens      int
	Cost        float64
	IsError     bool
	ErrorMsg    string
	ToolCalls   []string
	HasThinking bool
}

// Config holds app configuration
type Config struct {
	OpenClawDir string
	Port        string
}

var config Config
var templates *template.Template
var cronJobNames map[string]string // cron job ID -> name

func main() {
	config = Config{
		OpenClawDir: getEnv("OPENCLAW_DIR", filepath.Join(os.Getenv("HOME"), ".openclaw")),
		Port:        getEnv("PORT", "3600"),
	}

	// Load cron job names
	cronJobNames = loadCronJobNames()

	funcMap := template.FuncMap{
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},
		"formatCost": func(c float64) string {
			return fmt.Sprintf("$%.2f", c)
		},
	}

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseFS(content, "templates/*.html")
	if err != nil {
		log.Fatal("Failed to parse templates:", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	staticFS, _ := fs.Sub(content, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/", handleIndex)
	r.Get("/sessions", handleSessions)
	r.Get("/partials/sessions", handleSessionsPartial)
	r.Get("/session/{id}", handleSessionDetail)
	r.Get("/api/sessions", handleAPIListSessions)
	r.Get("/api/session/{id}/transcript", handleAPITranscript)

	addr := ":" + config.Port
	log.Printf("📡 Antenna starting on http://localhost%s", addr)
	log.Printf("   OpenClaw dir: %s", config.OpenClawDir)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadCronJobNames() map[string]string {
	names := make(map[string]string)
	
	jobsFile := filepath.Join(config.OpenClawDir, "cron", "jobs.json")
	data, err := os.ReadFile(jobsFile)
	if err != nil {
		log.Printf("Could not load cron jobs: %v", err)
		return names
	}

	var jobs CronJobsFile
	if err := json.Unmarshal(data, &jobs); err != nil {
		log.Printf("Could not parse cron jobs: %v", err)
		return names
	}

	for _, job := range jobs.Jobs {
		names[job.ID] = job.Name
	}
	
	log.Printf("Loaded %d cron job names", len(names))
	return names
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/sessions", http.StatusFound)
}

func buildSessionData(sessions []Session) map[string]interface{} {
	var activeSessions []Session
	var idleSessions []Session
	var subAgents []Session
	var cronJobs []Session
	var totalCost, todayCost float64

	for _, s := range sessions {
		totalCost += s.TotalCost
		todayCost += s.TodayCost

		switch s.Kind {
		case "subagent":
			subAgents = append(subAgents, s)
		case "cron":
			cronJobs = append(cronJobs, s)
		default:
			if s.IsActive {
				activeSessions = append(activeSessions, s)
			} else {
				idleSessions = append(idleSessions, s)
			}
		}
	}

	return map[string]interface{}{
		"Sessions":       sessions,
		"ActiveSessions": activeSessions,
		"IdleSessions":   idleSessions,
		"SubAgents":      subAgents,
		"CronJobs":       cronJobs,
		"TotalCount":     len(sessions),
		"ActiveCount":    len(activeSessions),
		"SubCount":       len(subAgents),
		"CronCount":      len(cronJobs),
		"TotalCost":      totalCost,
		"TodayCost":      todayCost,
	}
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := loadSessions()
	data := buildSessionData(sessions)
	data["Title"] = "Dashboard"
	data["Page"] = "sessions"

	if err := templates.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), 500)
	}
}

func handleSessionsPartial(w http.ResponseWriter, r *http.Request) {
	sessions := loadSessions()
	data := buildSessionData(sessions)

	if err := templates.ExecuteTemplate(w, "dashboard_content", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), 500)
	}
}

func handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	session, messages := loadSessionDetail(sessionID)

	data := map[string]interface{}{
		"Session":  session,
		"Messages": messages,
		"Title":    "Session: " + sessionID[:8],
		"Page":     "session",
	}

	if err := templates.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), 500)
	}
}

func handleAPIListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := loadSessions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func handleAPITranscript(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	_, messages := loadSessionDetail(sessionID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func loadSessionsFromJSON() SessionsJSON {
	sessionsFile := filepath.Join(config.OpenClawDir, "agents", "main", "sessions", "sessions.json")
	data, err := os.ReadFile(sessionsFile)
	if err != nil {
		return nil
	}

	var sessions SessionsJSON
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil
	}
	return sessions
}

// parseSessionKind extracts kind from session key - this is the authoritative source
func parseSessionKind(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		switch parts[2] {
		case "cron":
			return "cron"
		case "subagent":
			return "subagent"
		case "main":
			return "main"
		}
	}
	return "main"
}

// extractCronID gets the cron job ID from a session key like "agent:main:cron:abc123-..."
func extractCronID(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 4 && parts[2] == "cron" {
		return parts[3]
	}
	return ""
}

func loadSessions() []Session {
	var sessions []Session
	seen := make(map[string]bool)

	sessionsData := loadSessionsFromJSON()

	// Build maps for quick lookup
	sessionMeta := make(map[string]struct {
		Key   string
		Entry SessionEntry
	})
	keyBySessionID := make(map[string]string)
	
	if sessionsData != nil {
		for key, entry := range sessionsData {
			sessionMeta[entry.SessionID] = struct {
				Key   string
				Entry SessionEntry
			}{key, entry}
			keyBySessionID[entry.SessionID] = key
		}
	}

	sessionsDir := filepath.Join(config.OpenClawDir, "agents", "main", "sessions")
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		log.Printf("Failed to read sessions dir: %v", err)
		return sessions
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".jsonl") {
			continue
		}

		sessionID := strings.TrimSuffix(f.Name(), ".jsonl")
		if seen[sessionID] {
			continue
		}
		seen[sessionID] = true

		info, _ := f.Info()

		session := Session{
			SessionID:   sessionID,
			UpdatedTime: info.ModTime(),
			IsActive:    time.Since(info.ModTime()) < 30*time.Minute,
			Kind:        "main", // default
		}

		// Check if we have metadata from sessions.json - this is authoritative for type
		if meta, ok := sessionMeta[sessionID]; ok {
			session.Key = meta.Key
			session.Name = meta.Entry.Label
			session.Model = meta.Entry.Model
			session.Channel = meta.Entry.Channel
			session.TotalTokens = meta.Entry.TotalTokens
			session.Kind = parseSessionKind(meta.Key) // Use key-based detection
			
			if meta.Entry.UpdatedAt > 0 {
				session.UpdatedTime = time.UnixMilli(meta.Entry.UpdatedAt)
				session.IsActive = time.Since(session.UpdatedTime) < 30*time.Minute
			}
		}

		// Derive display name
		if session.Name == "" {
			switch session.Kind {
			case "main":
				// For main sessions, show date
				session.Name = session.UpdatedTime.Format("Jan 2 15:04")
			case "cron":
				// Try to get actual cron job name
				cronID := extractCronID(session.Key)
				if cronID != "" {
					if name, ok := cronJobNames[cronID]; ok {
						session.Name = name
					} else {
						session.Name = "cron-" + cronID[:8]
					}
				} else {
					session.Name = "cron-" + sessionID[:8]
				}
			case "subagent":
				session.Name = sessionID[:12]
			default:
				session.Name = sessionID[:12]
			}
		}

		// Set badge and color
		switch session.Kind {
		case "subagent":
			session.TypeBadge = "SUB"
			session.TypeColor = "purple"
		case "cron":
			session.TypeBadge = "CRON"
			session.TypeColor = "orange"
		default:
			session.TypeBadge = "MAIN"
			session.TypeColor = "cyan"
		}

		// Parse transcript for cost/message data
		parseSessionCostWithToday(sessionID, &session, today)

		sessions = append(sessions, session)
	}

	// Sort by updated time, most recent first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedTime.After(sessions[j].UpdatedTime)
	})

	return sessions
}

func parseSessionCostWithToday(sessionID string, session *Session, today time.Time) {
	transcriptPath := filepath.Join(config.OpenClawDir, "agents", "main", "sessions", sessionID+".jsonl")

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Type == "session" && entry.CWD != "" {
			session.CWD = entry.CWD
		}

		if entry.Type == "message" && entry.Message != nil {
			session.MessageCount++
			if entry.Message.Usage != nil && entry.Message.Usage.Cost != nil {
				cost := entry.Message.Usage.Cost.Total
				session.TotalCost += cost

				// Check if this message is from today
				var msgTime time.Time
				if entry.Message.Timestamp > 0 {
					msgTime = time.UnixMilli(entry.Message.Timestamp)
				} else if entry.Timestamp != "" {
					msgTime, _ = time.Parse(time.RFC3339, entry.Timestamp)
				}
				
				if !msgTime.IsZero() && msgTime.After(today) {
					session.TodayCost += cost
				}
			}
		}
	}
}

func loadSessionDetail(sessionID string) (*Session, []DisplayMessage) {
	session := &Session{SessionID: sessionID}
	var messages []DisplayMessage

	sessionsData := loadSessionsFromJSON()
	for key, entry := range sessionsData {
		if entry.SessionID == sessionID {
			session.Key = key
			session.Name = entry.Label
			session.Kind = parseSessionKind(key)
			session.Model = entry.Model
			session.TotalTokens = entry.TotalTokens
			
			// Get cron name if applicable
			if session.Kind == "cron" && session.Name == "" {
				cronID := extractCronID(key)
				if name, ok := cronJobNames[cronID]; ok {
					session.Name = name
				}
			}
			break
		}
	}

	transcriptPath := filepath.Join(config.OpenClawDir, "agents", "main", "sessions", sessionID+".jsonl")

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		log.Printf("Failed to read transcript: %v", err)
		return session, messages
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Type == "session" && entry.CWD != "" {
			session.CWD = entry.CWD
		}
		if entry.Type == "model_change" && entry.ModelID != "" {
			if session.Model == "" {
				session.Model = entry.ModelID
			}
			session.Provider = entry.Provider
		}

		if entry.Type == "message" && entry.Message != nil {
			msg := DisplayMessage{
				Role:  entry.Message.Role,
				Model: entry.Message.Model,
			}

			if entry.Message.Timestamp > 0 {
				msg.Time = time.UnixMilli(entry.Message.Timestamp)
			} else if entry.Timestamp != "" {
				msg.Time, _ = time.Parse(time.RFC3339, entry.Timestamp)
			}

			if entry.Message.Usage != nil {
				msg.Tokens = entry.Message.Usage.TotalTokens
				if entry.Message.Usage.Cost != nil {
					msg.Cost = entry.Message.Usage.Cost.Total
				}
				session.TotalCost += msg.Cost
			}

			if entry.Message.StopReason == "error" {
				msg.IsError = true
				msg.ErrorMsg = entry.Message.ErrorMessage
			}

			msg.Content, msg.ToolCalls, msg.HasThinking = parseContent(entry.Message.Content)

			session.MessageCount++
			messages = append(messages, msg)
		}
	}

	if session.Kind == "" {
		session.Kind = "main"
	}
	switch session.Kind {
	case "subagent":
		session.TypeBadge = "SUB"
		session.TypeColor = "purple"
	case "cron":
		session.TypeBadge = "CRON"
		session.TypeColor = "orange"
	default:
		session.TypeBadge = "MAIN"
		session.TypeColor = "cyan"
	}

	return session, messages
}

func parseContent(raw json.RawMessage) (text string, toolCalls []string, hasThinking bool) {
	if len(raw) == 0 {
		return "", nil, false
	}

	var str string
	if json.Unmarshal(raw, &str) == nil {
		return str, nil, false
	}

	var blocks []ContentBlock
	if json.Unmarshal(raw, &blocks) == nil {
		var texts []string
		for _, block := range blocks {
			switch block.Type {
			case "text":
				texts = append(texts, block.Text)
			case "toolCall":
				toolCalls = append(toolCalls, block.Name)
			case "thinking":
				hasThinking = true
			}
		}
		return strings.Join(texts, "\n"), toolCalls, hasThinking
	}

	return string(raw), nil, false
}
