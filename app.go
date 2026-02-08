package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// App struct
type App struct {
	ctx         context.Context
	openclawDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		openclawDir: filepath.Join(os.Getenv("HOME"), ".openclaw"),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Session represents a session for the frontend
type Session struct {
	SessionID    string  `json:"sessionId"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Model        string  `json:"model"`
	MessageCount int     `json:"messageCount"`
	TotalCost    float64 `json:"totalCost"`
	TodayCost    float64 `json:"todayCost"`
	UpdatedAt    int64   `json:"updatedAt"`
	IsActive     bool    `json:"isActive"`
}

// DashboardData is the full dashboard response
type DashboardData struct {
	Sessions   []Session `json:"sessions"`
	TotalCount int       `json:"totalCount"`
	TotalCost  float64   `json:"totalCost"`
	TodayCost  float64   `json:"todayCost"`
}

// SessionsJSON from sessions.json
type sessionsJSON map[string]sessionEntry

type sessionEntry struct {
	SessionID   string `json:"sessionId"`
	UpdatedAt   int64  `json:"updatedAt"`
	Label       string `json:"label,omitempty"`
	Model       string `json:"model,omitempty"`
	TotalTokens int    `json:"totalTokens"`
}

type cronJobsFile struct {
	Jobs []cronJob `json:"jobs"`
}

type cronJob struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type transcriptEntry struct {
	Type    string          `json:"type"`
	Message *messageContent `json:"message,omitempty"`
}

type messageContent struct {
	Timestamp int64      `json:"timestamp,omitempty"`
	Usage     *usageInfo `json:"usage,omitempty"`
}

type usageInfo struct {
	Cost *costInfo `json:"cost,omitempty"`
}

type costInfo struct {
	Total float64 `json:"total"`
}

// GetDashboard returns the dashboard data
func (a *App) GetDashboard() DashboardData {
	sessions := a.loadSessions()
	
	var totalCost, todayCost float64
	for _, s := range sessions {
		totalCost += s.TotalCost
		todayCost += s.TodayCost
	}
	
	return DashboardData{
		Sessions:   sessions,
		TotalCount: len(sessions),
		TotalCost:  totalCost,
		TodayCost:  todayCost,
	}
}

func (a *App) loadCronJobNames() map[string]string {
	names := make(map[string]string)
	jobsFile := filepath.Join(a.openclawDir, "cron", "jobs.json")
	data, err := os.ReadFile(jobsFile)
	if err != nil {
		return names
	}
	var jobs cronJobsFile
	if err := json.Unmarshal(data, &jobs); err != nil {
		return names
	}
	for _, job := range jobs.Jobs {
		names[job.ID] = job.Name
	}
	return names
}

func (a *App) loadSessions() []Session {
	var sessions []Session
	cronNames := a.loadCronJobNames()
	
	// Load sessions.json for metadata
	sessionsFile := filepath.Join(a.openclawDir, "agents", "main", "sessions", "sessions.json")
	var sessionMeta sessionsJSON
	if data, err := os.ReadFile(sessionsFile); err == nil {
		json.Unmarshal(data, &sessionMeta)
	}
	
	// Build lookup
	metaByID := make(map[string]struct {
		Key   string
		Entry sessionEntry
	})
	for key, entry := range sessionMeta {
		metaByID[entry.SessionID] = struct {
			Key   string
			Entry sessionEntry
		}{key, entry}
	}
	
	// Scan session files
	sessionsDir := filepath.Join(a.openclawDir, "agents", "main", "sessions")
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return sessions
	}
	
	today := time.Now().Truncate(24 * time.Hour)
	seen := make(map[string]bool)
	
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
		
		s := Session{
			SessionID: sessionID,
			Kind:      "main",
			UpdatedAt: info.ModTime().UnixMilli(),
			IsActive:  time.Since(info.ModTime()) < 30*time.Minute,
		}
		
		// Get metadata
		if meta, ok := metaByID[sessionID]; ok {
			s.Name = meta.Entry.Label
			s.Model = meta.Entry.Model
			s.Kind = parseKind(meta.Key)
			if meta.Entry.UpdatedAt > 0 {
				s.UpdatedAt = meta.Entry.UpdatedAt
				s.IsActive = time.Since(time.UnixMilli(meta.Entry.UpdatedAt)) < 30*time.Minute
			}
			
			// Get cron name
			if s.Kind == "cron" && s.Name == "" {
				parts := strings.Split(meta.Key, ":")
				if len(parts) >= 4 {
					if name, ok := cronNames[parts[3]]; ok {
						s.Name = name
					}
				}
			}
		}
		
		// Default name
		if s.Name == "" {
			switch s.Kind {
			case "main":
				s.Name = time.UnixMilli(s.UpdatedAt).Format("Jan 2 15:04")
			case "cron":
				s.Name = "cron-" + sessionID[:8]
			default:
				s.Name = sessionID[:12]
			}
		}
		
		// Parse transcript for costs
		a.parseSessionCost(sessionID, &s, today)
		
		sessions = append(sessions, s)
	}
	
	// Sort by updated time
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})
	
	return sessions
}

func parseKind(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 3 {
		switch parts[2] {
		case "cron":
			return "cron"
		case "subagent":
			return "subagent"
		}
	}
	return "main"
}

// HourlyBucket represents activity in one hour
type HourlyBucket struct {
	Hour     string  `json:"hour"`     // "HH:00" format
	Messages int     `json:"messages"`
	Cost     float64 `json:"cost"`
}

// GetHourlyActivity returns message counts and costs bucketed by hour for the last 24h
func (a *App) GetHourlyActivity() []HourlyBucket {
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	// Initialize 24 buckets
	buckets := make([]HourlyBucket, 24)
	for i := 0; i < 24; i++ {
		t := cutoff.Add(time.Duration(i+1) * time.Hour)
		buckets[i] = HourlyBucket{Hour: t.Format("15:00")}
	}

	sessionsDir := filepath.Join(a.openclawDir, "agents", "main", "sessions")
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return buckets
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, f.Name()))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}
			var entry transcriptEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.Type != "message" || entry.Message == nil || entry.Message.Timestamp <= 0 {
				continue
			}
			msgTime := time.UnixMilli(entry.Message.Timestamp)
			if msgTime.Before(cutoff) || msgTime.After(now) {
				continue
			}
			// Find bucket index
			hoursSinceCutoff := msgTime.Sub(cutoff).Hours()
			idx := int(hoursSinceCutoff)
			if idx >= 24 {
				idx = 23
			}
			buckets[idx].Messages++
			if entry.Message.Usage != nil && entry.Message.Usage.Cost != nil {
				buckets[idx].Cost += entry.Message.Usage.Cost.Total
			}
		}
	}

	return buckets
}

func (a *App) parseSessionCost(sessionID string, s *Session, today time.Time) {
	path := filepath.Join(a.openclawDir, "agents", "main", "sessions", sessionID+".jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var entry transcriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Type == "message" && entry.Message != nil {
			s.MessageCount++
			if entry.Message.Usage != nil && entry.Message.Usage.Cost != nil {
				cost := entry.Message.Usage.Cost.Total
				s.TotalCost += cost
				if entry.Message.Timestamp > 0 {
					msgTime := time.UnixMilli(entry.Message.Timestamp)
					if msgTime.After(today) {
						s.TodayCost += cost
					}
				}
			}
		}
	}
}
