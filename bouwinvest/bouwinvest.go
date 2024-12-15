package bouwinvest

import (
	"fmt"
	"huurwoning/browser"
	"huurwoning/logger"
	"huurwoning/scraper"
	"strings"
	"sync"

	"github.com/chromedp/chromedp"
)

func BouwInvest(b *browser.Browser, globalLogger *logger.GlobalLogger, url string, prevResults map[string]struct{}, multipleResultsFromPreviousRun *[]string) error {
	logger := globalLogger.Logger("BOUWINVEST")

	scraper, err := scraper.NewScraper(b, "BOUWINVEST", url, "", prevResults, false, logger, GetResultsFactory, false, multipleResultsFromPreviousRun)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %v", err)
	}
	defer scraper.Close()

	scraper.Logger.Info("Start")

	newResults, err := scraper.GetResults(scraper, b, prevResults)
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

func GetResults(scraper *scraper.Scraper, b *browser.Browser, prevResults map[string]struct{}) (map[string]struct{}, error) {
	allResults := make(map[string]struct{})

	var mu sync.Mutex

	maxPages := 10

	for page := 1; page <= maxPages; page++ {
		err := processPage(scraper, b, page, &mu, allResults)
		if err != nil {
			return prevResults, err
		}
	}

	return allResults, nil
}

func processPage(scraper *scraper.Scraper, b *browser.Browser, page int, mu *sync.Mutex, allResults map[string]struct{}) error {
	pageUrl := fmt.Sprintf("%s&page=%d", scraper.Url, page)
	scraper.Logger.Info(fmt.Sprintf("Visit page %d", page))

	var results []string

	// Create tasks to navigate to the URL and extract text for all elements matching the selector
	err := b.RunInTab(scraper.TabCtx,
		chromedp.Navigate(pageUrl),
		chromedp.WaitVisible("span.h2.fw-book.color-orange"),
		chromedp.Evaluate(fmt.Sprintf(`
            Array.from(document.querySelectorAll('%s')).map(el => el.textContent)
        `, "span.h2.fw-book.color-orange"), &results),
	)

	if err != nil {
		scraper.Logger.Error(fmt.Sprintf("Error opening website for page %d", page), "error", err)
		scraper.HasError = true
		return err
	}

	scraper.Logger.Info(fmt.Sprintf("Found multiple results on page %d: %v", page, strings.Join(results, ", ")))

	mu.Lock()
	for _, result := range results {
		if result == "" {
			continue
		}
		trimmed := strings.TrimSpace(result)
		allResults[trimmed] = struct{}{}
	}
	mu.Unlock()
	return nil
}
