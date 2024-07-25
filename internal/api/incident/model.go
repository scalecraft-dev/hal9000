package incident

type CreateIncidentRequest struct {
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
