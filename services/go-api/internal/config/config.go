package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv            string
	Port              string
	MySQLDSN          string
	SessionSecret     string
	SessionTTLHours   int
	MiniProgramAppID  string
	MiniProgramSecret string
	WeChatMockLogin   bool
	CORSAllowOrigins  []string
}

func Load() Config {
	loadDotEnv(".env")
	ttl, _ := strconv.Atoi(getenv("SESSION_TTL_HOURS", "720"))
	if ttl <= 0 {
		ttl = 720
	}
	return Config{
		AppEnv:            getenv("APP_ENV", "development"),
		Port:              getenv("PORT", "10240"),
		MySQLDSN:          getenv("MYSQL_DSN", "accounting:accounting@tcp(127.0.0.1:3306)/accounting?charset=utf8mb4&parseTime=True&loc=Local"),
		SessionSecret:     getenv("SESSION_TOKEN_SECRET", "change-me-in-production"),
		SessionTTLHours:   ttl,
		MiniProgramAppID:  getenv("MINIPROGRAM_APPID", ""),
		MiniProgramSecret: getenv("MINIPROGRAM_SECRET", ""),
		WeChatMockLogin:   strings.EqualFold(getenv("WECHAT_MOCK_LOGIN", "false"), "true"),
		CORSAllowOrigins:  splitCSV(getenv("CORS_ALLOW_ORIGINS", "*")),
	}
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	if len(items) == 0 {
		return []string{"*"}
	}
	return items
}
