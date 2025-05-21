package internal

import (
	"time"
)

type Incident struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Severity    Severity  `json:"severity"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ChannelID   string    `json:"channel_id"`
	Members     []string  `json:"members"`
}

type Status string

const (
	StatusInvestigating Status = "Investigating"
	StatusFixing        Status = "Fixing"
	StatusMonitoring    Status = "Monitoring"
	StatusResolved      Status = "Resolved"
)

type Severity string

const (
	// Critical incident with high impact for a larger or total number of users
	SeveritySev0 Severity = "SEV-0"
	// Major incident with significant impact. Partial system disruption that affects a smaller number of users
	SeveritySev1 Severity = "SEV-1"
	// Less severe incident impacting non-critical functionalities or inconvenience for users
	SeveritySev2 Severity = "SEV-2"
	// Minor to low level incident for non critical feature areas or low priority user complaints.
	SeveritySev3 Severity = "SEV-3"
)

type TimelineItem struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	User      string    `json:"user"`
}

type ActionItem struct {
	Description string `json:"description"`
	User        string `json:"user"`
	Completed   bool   `json:"completed"`
}

type SlackCommandRequest struct {
	Token               string `form:"token" json:"token"`
	TeamId              string `form:"team_id" json:"team_id"`
	TeamDomain          string `form:"team_domain" json:"team_domain"`
	ChannelId           string `form:"channel_id" json:"channel_id"`
	ChannelName         string `form:"channel_name" json:"channel_name"`
	UserId              string `form:"user_id" json:"user_id"`
	UserName            string `form:"user_name" json:"user_name"`
	Command             string `form:"command" json:"command"`
	Text                string `form:"text" json:"text"`
	ApiAppId            string `form:"api_app_id" json:"api_app_id"`
	IsEnterpriseInstall bool   `form:"is_enterprise_install" json:"is_enterprise_install"`
	ResponseUrl         string `form:"response_url" json:"response_url"`
	TriggerId           string `form:"trigger_id" json:"trigger_id"`
}

type Config struct {
	SlackToken         string
	SlackSigningSecret string
	ServerPort         int
	ServerHost         string
	Environment        string
	LogLevel           string
}
