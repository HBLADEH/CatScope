package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"catscope/internal/ai"
	"catscope/internal/logcat"
)

const (
	SessionVersion   = 1
	SessionExtension = ".catscope-session"
)

type SessionFilters struct {
	Level          []string `json:"level"`
	PackageName    string   `json:"packageName"`
	Keyword        string   `json:"keyword"`
	RegexEnabled   bool     `json:"regexEnabled"`
	Tags           []string `json:"tags"`
	ExcludeKeyword string   `json:"excludeKeyword"`
}

type Session struct {
	Version          int                     `json:"version"`
	SessionID        string                  `json:"sessionId"`
	Name             string                  `json:"name"`
	CreatedAt        string                  `json:"createdAt"`
	UpdatedAt        string                  `json:"updatedAt"`
	SourceMode       string                  `json:"sourceMode"`
	SourceName       string                  `json:"sourceName"`
	SourcePath       string                  `json:"sourcePath"`
	WorkspaceID      string                  `json:"workspaceId"`
	WorkspaceName    string                  `json:"workspaceName"`
	ProjectPath      string                  `json:"projectPath"`
	PackageName      string                  `json:"packageName"`
	SelectedDevice   string                  `json:"selectedDevice"`
	KnownPIDs        []int                   `json:"knownPids"`
	Filters          SessionFilters          `json:"filters"`
	AIContextOptions ai.AIContextOptions     `json:"aiContextOptions"`
	LogEntries       []logcat.LogEntry       `json:"logEntries"`
	AnalysisResults  []logcat.AnalysisResult `json:"analysisResults"`
	Notes            string                  `json:"notes"`
}

type SessionSummary struct {
	SessionID     string `json:"sessionId"`
	Name          string `json:"name"`
	FilePath      string `json:"filePath"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	SourceMode    string `json:"sourceMode"`
	SourceName    string `json:"sourceName"`
	WorkspaceName string `json:"workspaceName"`
	PackageName   string `json:"packageName"`
	LogCount      int    `json:"logCount"`
	AnalysisCount int    `json:"analysisCount"`
}

func NewSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func NormalizeSession(session Session) Session {
	now := time.Now().UTC().Format(time.RFC3339)
	if session.Version <= 0 {
		session.Version = SessionVersion
	}
	session.SessionID = strings.TrimSpace(session.SessionID)
	if session.SessionID == "" {
		session.SessionID = NewSessionID()
	}
	session.Name = strings.TrimSpace(session.Name)
	if session.Name == "" {
		session.Name = "CatScope Session"
	}
	if strings.TrimSpace(session.CreatedAt) == "" {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	session.SourceMode = strings.TrimSpace(session.SourceMode)
	if session.SourceMode == "" {
		session.SourceMode = logcat.LogSourceLive
	}
	session.SourceName = strings.TrimSpace(session.SourceName)
	session.SourcePath = strings.TrimSpace(session.SourcePath)
	session.WorkspaceID = strings.TrimSpace(session.WorkspaceID)
	session.WorkspaceName = strings.TrimSpace(session.WorkspaceName)
	session.ProjectPath = strings.TrimSpace(session.ProjectPath)
	session.PackageName = strings.TrimSpace(session.PackageName)
	session.SelectedDevice = strings.TrimSpace(session.SelectedDevice)
	session.Filters.PackageName = strings.TrimSpace(session.Filters.PackageName)
	session.Filters.Keyword = strings.TrimSpace(session.Filters.Keyword)
	session.Filters.ExcludeKeyword = strings.TrimSpace(session.Filters.ExcludeKeyword)
	for index := range session.Filters.Tags {
		session.Filters.Tags[index] = strings.TrimSpace(session.Filters.Tags[index])
	}
	if session.AIContextOptions.Language == "" {
		session.AIContextOptions = ai.DefaultOptions()
	}
	return session
}

func SaveSession(path string, session Session) (Session, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Session{}, errors.New("session path is required")
	}
	if filepath.Ext(path) == "" {
		path += SessionExtension
	}
	if !strings.EqualFold(filepath.Ext(path), SessionExtension) {
		path += SessionExtension
	}
	session = NormalizeSession(session)
	if len(session.LogEntries) == 0 {
		return Session{}, errors.New("no logs to save")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Session{}, err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return Session{}, fmt.Errorf("encode session failed: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return Session{}, fmt.Errorf("write session failed: %w", err)
	}
	return session, nil
}

func OpenSession(path string) (Session, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Session{}, errors.New("session path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, fmt.Errorf("read session failed: %w", err)
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("invalid session file: %w", err)
	}
	if strings.TrimSpace(session.SessionID) == "" && len(session.LogEntries) == 0 {
		return Session{}, errors.New("invalid session file: missing session data")
	}
	session = NormalizeOpenedSession(session)
	if len(session.LogEntries) == 0 {
		return Session{}, errors.New("invalid session file: no log entries")
	}
	return session, nil
}

func NormalizeOpenedSession(session Session) Session {
	if session.Version <= 0 {
		session.Version = SessionVersion
	}
	if strings.TrimSpace(session.SessionID) == "" {
		session.SessionID = NewSessionID()
	}
	if strings.TrimSpace(session.Name) == "" {
		session.Name = "CatScope Session"
	}
	if strings.TrimSpace(session.SourceMode) == "" {
		session.SourceMode = logcat.LogSourceOffline
	}
	if strings.TrimSpace(session.CreatedAt) == "" {
		session.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(session.UpdatedAt) == "" {
		session.UpdatedAt = session.CreatedAt
	}
	if session.AIContextOptions.Language == "" {
		session.AIContextOptions = ai.DefaultOptions()
	}
	return session
}

func Summary(session Session, filePath string) SessionSummary {
	return SessionSummary{
		SessionID:     session.SessionID,
		Name:          session.Name,
		FilePath:      filePath,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
		SourceMode:    session.SourceMode,
		SourceName:    session.SourceName,
		WorkspaceName: session.WorkspaceName,
		PackageName:   session.PackageName,
		LogCount:      len(session.LogEntries),
		AnalysisCount: len(session.AnalysisResults),
	}
}
