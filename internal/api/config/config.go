package config

type Config struct {
	SlackToken         string
	SlackSigningSecret string
}

func GetConfig() *Config {
	c := Config{
		SlackToken:         getEnv("SLACK_TOKEN", "", true),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", "", true),
	}

	return &c
}
