package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config содержит конфигурационные данные приложения
type Config struct {
	Port          string
	Env           string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBSslMode     string
	EncryptionKey string
	JWTSecret     string
	GreenAPIURL           string
	GreenAPIIDInstance    string
	GreenAPITokenInstance string
}

// LoadConfig загружает переменные окружения из .env и системных переменных
func LoadConfig() (*Config, error) {
	// Загружаем .env файл, если он существует (полезно для локальной разработки)
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}

	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPortStr := getEnv("DB_PORT", "5432")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		dbPort = 5432
	}

	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "dapter_db")
	dbSslMode := getEnv("DB_SSLMODE", "disable")

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Println("ВНИМАНИЕ: ENCRYPTION_KEY не установлен. Безопасность данных под угрозой!")
	}

	jwtSecret := getEnv("JWT_SECRET", "super_secret_jwt_key_change_me_in_production")
	greenApiUrl := getEnv("GREEN_API_URL", "https://api.green-api.com")
	greenApiIdInstance := getEnv("GREEN_API_ID_INSTANCE", "")
	greenApiTokenInstance := getEnv("GREEN_API_TOKEN_INSTANCE", "")

	return &Config{
		Port:          port,
		Env:           env,
		DBHost:        dbHost,
		DBPort:        dbPort,
		DBUser:        dbUser,
		DBPassword:    dbPassword,
		DBName:        dbName,
		DBSslMode:     dbSslMode,
		EncryptionKey: encryptionKey,
		JWTSecret:     jwtSecret,
		GreenAPIURL:           greenApiUrl,
		GreenAPIIDInstance:    greenApiIdInstance,
		GreenAPITokenInstance: greenApiTokenInstance,
	}, nil
}

// getEnv вспомогательная функция для получения значения или дефолта
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
