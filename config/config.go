package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ENVIRONMENT string
	DEBUG_MODE  bool

	DB_PATH string

	USER_NAME  string
	REBO_PW    string
	BEUMER_PW  string
	VESTEDA_PW string

	TWILIO_SID          string
	TWILIO_TOKEN        string
	TWILIO_PHONE_NUMBER string
	YOUR_PHONE_NUMBER   string
	SMTP_SERVER         string
	SMTP_PORT           int
	SMTP_USERNAME       string
	SMTP_PASSWORD       string
	TO_EMAIL            string
	FROM_EMAIL          string

	GRAFANA_USERNAME string
	GRAFANA_PASSWORD string
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatalf("Failed to convert SMTP_PORT to int: %v", err)
	}

	config := &Config{
		ENVIRONMENT: os.Getenv("ENVIRONMENT"),
		DEBUG_MODE:  os.Getenv("DEBUG_MODE") == "true",

		DB_PATH: os.Getenv("DB_PATH"),

		USER_NAME:  os.Getenv("USER_NAME"),
		REBO_PW:    os.Getenv("REBO_PW"),
		BEUMER_PW:  os.Getenv("BEUMER_PW"),
		VESTEDA_PW: os.Getenv("VESTEDA_PW"),

		TWILIO_SID:          os.Getenv("TWILIO_SID"),
		TWILIO_TOKEN:        os.Getenv("TWILIO_TOKEN"),
		TWILIO_PHONE_NUMBER: os.Getenv("TWILIO_PHONE_NUMBER"),
		YOUR_PHONE_NUMBER:   os.Getenv("YOUR_PHONE_NUMBER"),
		SMTP_SERVER:         os.Getenv("SMTP_SERVER"),
		SMTP_PORT:           port,
		SMTP_USERNAME:       os.Getenv("SMTP_USERNAME"),
		SMTP_PASSWORD:       os.Getenv("SMTP_PASSWORD"),
		TO_EMAIL:            os.Getenv("TO_EMAIL"),
		FROM_EMAIL:          os.Getenv("FROM_EMAIL"),

		GRAFANA_USERNAME: os.Getenv("GRAFANA_USERNAME"),
		GRAFANA_PASSWORD: os.Getenv("GRAFANA_PASSWORD"),
	}

	return config, nil
}
