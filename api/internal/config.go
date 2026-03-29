package notifystock

import (
	"fmt"
	"log/slog"
	"os"

	yaml "github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
)

var Cfg Config

func init() {
	if err := initConfig(); err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		panic(err) // initではpanicが必要だが、エラー情報を改善
	}
}

func initConfig() error {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			slog.Info("No .env file found, using environment variables")
		}
	}

	cfg, err := loadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	Cfg = *cfg
	logger = CreateLogger(Cfg.LogLevel)
	return nil
}

func loadConfigFromEnv() (*Config, error) {
	requiredEnvs := map[string]string{
		"FROM":                "",
		"TO":                  "",
		"MAIL_TOKEN":          "",
		"OAUTH_CLIENT_ID":     "",
		"OAUTH_CLIENT_SECRET": "",
		"OAUTH_REDIRECT_URL":  "",
		"FRONTEND_URL":        "",
		"MAIL_GUN_API_KEY":    "",
		"MAIL_DOMAIN":         "",
	}

	// 必須環境変数の確認
	for key := range requiredEnvs {
		value, ok := os.LookupEnv(key)
		if !ok {
			return nil, fmt.Errorf("required environment variable %s is not set", key)
		}
		requiredEnvs[key] = value
	}

	// データベース環境変数（デフォルト値付き）
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5555")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "notify-stock")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	// 後方互換性のため、DBDSNが設定されている場合はそれを優先
	dbdsn, ok := os.LookupEnv("DBDSN")
	if !ok {
		dbdsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		FROM:              requiredEnvs["FROM"],
		TO:                requiredEnvs["TO"],
		DBDSN:             dbdsn,
		DBHost:            dbHost,
		DBPort:            dbPort,
		DBUser:            dbUser,
		DBPassword:        dbPassword,
		DBName:            dbName,
		DBSSLMode:         dbSSLMode,
		MailToken:         requiredEnvs["MAIL_TOKEN"],
		MailDomain:        requiredEnvs["MAIL_DOMAIN"],
		MailGunAPIKey:     requiredEnvs["MAIL_GUN_API_KEY"],
		OauthClientID:     requiredEnvs["OAUTH_CLIENT_ID"],
		OauthClientSecret: requiredEnvs["OAUTH_CLIENT_SECRET"],
		OauthRedirectURL:  requiredEnvs["OAUTH_REDIRECT_URL"],
		FrontendURL:       requiredEnvs["FRONTEND_URL"],
		LogLevel:          parseLogLevel(logLevel),
		Environment:       env,
	}, nil
}

type Config struct {
	FROM              string
	TO                string
	DBDSN             string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	MailToken         string
	MailGunAPIKey     string
	MailDomain        string
	OauthClientID     string
	OauthClientSecret string
	OauthRedirectURL  string
	FrontendURL       string // フロントエンドのURLを追加
	LogLevel          slog.Level
	Environment       string
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

type SupportSymbol struct {
	Symbols []string `yaml:"symbols"`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetSupportSymbols(path string) (*SupportSymbol, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var symbol SupportSymbol
	if err := yaml.Unmarshal(buf, &symbol); err != nil {
		return nil, err
	}
	return &symbol, nil
}
