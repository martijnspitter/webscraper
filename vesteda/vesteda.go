package vesteda

import (
	"fmt"
	"huurwoning/browser"
	"huurwoning/config"
	"huurwoning/db"
	"huurwoning/logger"
	"huurwoning/scraper"
	"strings"

	"github.com/chromedp/chromedp"
)

func Vesteda(b *browser.Browser, globalLogger *logger.GlobalLogger, url string, db *db.Database) error {
	logger := globalLogger.Logger("VESTEDA")

	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	scraper, err := scraper.NewScraper(b, "VESTEDA", url, config.VESTEDA_PW, false, logger, GetResultsFactory, config.DEBUG_MODE, db)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %v", err)
	}
	defer scraper.Close()

	scraper.Logger.Info("Start")

	// Login if needed
	err = scraper.LoginIfNeeded(b)
	if err != nil {
		return err
	}

	newResults, err := scraper.GetResults(scraper, b)
	if err != nil {
		scraper.CheckForNewResults(newResults)
		scraper.UpdatePrevResults(newResults)
		return err
	}

	scraper.CheckForNewResults(newResults)
	scraper.UpdatePrevResults(newResults)
	return nil
}

func GetResultsFactory() scraper.GetResults {
	return GetResults
}

func GetResults(scraper *scraper.Scraper, b *browser.Browser) ([]string, error) {
	var results []string

	err := b.RunInTab(scraper.TabCtx,
		chromedp.Navigate("https://hurenbij.vesteda.com/zoekopdracht/"),
		chromedp.WaitVisible("a.stretched-link", chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('a.stretched-link')).map(el => el.textContent.trim())
		`, &results),
	)

	if err != nil {
		scraper.Logger.Error("Error getting results", err)
		scraper.HasError = true
		return results, fmt.Errorf("Error getting results: %v", err)
	}

	// Create a map to store the current results
	currentResults := make([]string, 0, len(results))

	// Iterate over the elements and store their text in the current results map
	for _, result := range results {
		if result == "" {
			continue
		}
		trimmed := strings.TrimSpace(result)
		currentResults = append(currentResults, trimmed)
	}

	return currentResults, nil
}
