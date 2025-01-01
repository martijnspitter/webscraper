package main

import (
	"log"
	"time"

	"huurwoning/beumer"
	"huurwoning/bouwinvest"
	"huurwoning/browser"
	"huurwoning/config"
	"huurwoning/db"
	"huurwoning/logger"
	"huurwoning/rebo"
	"huurwoning/vesteda"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	globalLogger, err := logger.NewGlobalLogger(config.ENVIRONMENT == "development")
	if err != nil {
		log.Fatalf("Failed to create global logger: %v", err)
	}
	defer globalLogger.Close()

	dbPath := config.DB_PATH
	if config.ENVIRONMENT == "development" {
		dbPath = "./data/properties.db"
	}

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	logger := globalLogger.Logger("MAIN")

	b, err := browser.New(config.DEBUG_MODE, globalLogger)
	if err != nil {
		log.Fatalf("Failed to create browser: %v", err)
	}
	defer b.Close()

	// Start separate timers for each scraping process
	for {
		err := rebo.Rebo(b, globalLogger, "https://rebowonenhuur.nl/login", database)
		if err != nil {
			logger.Error("Error in Rebo scraping!", "error", err)
		}

		err = vesteda.Vesteda(b, globalLogger, "https://hurenbij.vesteda.com/login", database)
		if err != nil {
			logger.Error("Error in Vesteda scraping!", "error", err)
		}

		err = bouwinvest.BouwInvest(b, globalLogger, "https://www.wonenbijbouwinvest.nl/huuraanbod?query=Utrecht&range=10&seniorservice=false&order=recent&size=50", database)
		if err != nil {
			logger.Error("Error in BouwInvest scraping!", "error", err)
		}

		err = beumer.Beumer(b, globalLogger, "https://www.beumer.nl/huurwoningen/?search=Utrecht&status%5B0%5D=te-huur", database)
		if err != nil {
			logger.Error("Error in Beumer scraping!", "error", err)
		}

		time.Sleep(30 * time.Second)
	}
}
