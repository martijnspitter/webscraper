package scraper

import (
	"context"
	"fmt"
	"strings"

	"huurwoning/browser"
	"huurwoning/logger"
	"huurwoning/reporting"

	"github.com/chromedp/chromedp"
)

type GetResultsFactory func() GetResults
type GetResults func(s *Scraper, b *browser.Browser, prevResults map[string]struct{}) (map[string]struct{}, error)

type Scraper struct {
	name                           string
	Url                            string
	username                       string
	password                       string
	prevResults                    map[string]struct{}
	HasError                       bool
	Logger                         *logger.Logger
	GetResults                     GetResults
	isDebugging                    bool
	multipleResultsFromPreviousRun *[]string
	TabCtx                         context.Context
	tabCancel                      context.CancelFunc
	browser                        *browser.Browser
}

func (s *Scraper) UpdatePrevResults(newResults map[string]struct{}) {
	// unable to get all the new results, so we'll just return
	if s.HasError || len(*s.multipleResultsFromPreviousRun) > 0 || len(newResults) == 0 {
		return
	}
	for k := range s.prevResults {
		delete(s.prevResults, k)
	}
	for k, v := range newResults {
		s.prevResults[k] = v
	}
}

func (s *Scraper) CheckForNewResults(currentResults map[string]struct{}) {
	if len(currentResults) == 0 {
		s.Logger.Warn("No results found.")
		return
	}

	// Compare current results with previous results and log new results
	newResults := make([]string, 0)
	for text := range currentResults {
		if _, found := s.prevResults[text]; !found {
			newResults = append(newResults, text)
		}
	}

	// When we find multiple new results the first time dont send an alert
	if len(newResults) > 1 && len(*s.multipleResultsFromPreviousRun) == 0 {
		s.Logger.Warn("Found multiple results! Not alerting", "results", strings.Join(newResults, ", "))
		*s.multipleResultsFromPreviousRun = newResults
		return
	}

	switch len(newResults) {
	case 0:
		s.Logger.Info("No new results found.")
	case 1:
		s.Logger.Warn(fmt.Sprintf("New result found %s", newResults[0]))
		if !s.isDebugging {
			reporting.SendAlert(newResults[0], fmt.Sprintf("%s: ", s.name), s.Logger)
		}
	default:
		s.Logger.Warn(fmt.Sprintf("%d new results found.", len(newResults)))
		if !s.isDebugging {
			// Send one alert with all results
			allResults := strings.Join(newResults, "\n")
			reporting.SendAlertForMultipleResults(allResults, fmt.Sprintf("%s: Multiple Results\n", s.name), s.Logger)
		}
	}

	*s.multipleResultsFromPreviousRun = []string{}
}

func (s *Scraper) createTab() error {
	var err error
	s.TabCtx, s.tabCancel, err = s.browser.CreateTab()
	if err != nil {
		return fmt.Errorf("failed to create tab: %v", err)
	}
	return nil
}

func (s *Scraper) Close() {
	if s.tabCancel != nil {
		s.tabCancel()
		s.TabCtx = nil
		s.tabCancel = nil
		s.browser.DecreaseTabCount()
	}
}

func (s *Scraper) ensureBrowserAlive() error {
	if !s.browser.IsAlive() {
		s.Logger.Warn("Browser is not alive, recreating...")
		err := s.browser.RecreateIfNeeded()
		if err != nil {
			return fmt.Errorf("failed to recreate browser: %v", err)
		}
		// Recreate the tab for this scraper
		err = s.createTab()
		if err != nil {
			return fmt.Errorf("failed to recreate tab: %v", err)
		}
	}
	return nil
}

func (s *Scraper) LoginIfNeeded(b *browser.Browser) error {
	err := s.ensureBrowserAlive()
	if err != nil {
		return err
	}

	if s.TabCtx == nil {
		return fmt.Errorf("tab not created")
	}

	s.Logger.Info("Opening the login page...")

	var isLoggedIn bool
	err = b.RunInTab(s.TabCtx,
		chromedp.Navigate(s.Url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('div.alert')).some(
                alert => alert.innerText.includes('Welkom, je bent reeds ingelogd.')
            )
        `, &isLoggedIn),
	)
	if err != nil {
		return fmt.Errorf("error checking login status: %v", err)
	}

	if isLoggedIn {
		s.Logger.Info("Already logged in.")
		return nil
	}

	s.Logger.Info("Logging in...")

	err = b.RunInTab(s.TabCtx,
		chromedp.WaitVisible(`input[name="txtEmail"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="txtEmail"]`, s.username, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="txtWachtwoord"]`, s.password, chromedp.ByQuery),
		chromedp.Click(`//button[contains(text(), 'Inloggen')]`, chromedp.BySearch),
		chromedp.WaitNotPresent(`//button[contains(text(), 'Inloggen')]`, chromedp.BySearch),
	)
	if err != nil {
		return fmt.Errorf("error during login process: %v", err)
	}

	// Verify login success
	var loginSuccess bool
	err = b.RunInTab(s.TabCtx,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`document.body.innerText.includes('Welkom')`, &loginSuccess),
	)
	if err != nil {
		return fmt.Errorf("error verifying login success: %v", err)
	}

	if !loginSuccess {
		return fmt.Errorf("login failed: welcome message not found")
	}

	s.Logger.Info("Login Success!")
	return nil
}

func NewScraper(b *browser.Browser, name, url string, password string, prevResults map[string]struct{}, hasError bool, logger *logger.Logger, getResultsFactory GetResultsFactory, isDebugging bool, multipleResultsFromPreviousRun *[]string) (*Scraper, error) {
	s := &Scraper{
		name:                           name,
		Url:                            url,
		username:                       "martijnspitter@gmail.com",
		password:                       password,
		prevResults:                    prevResults,
		HasError:                       hasError,
		Logger:                         logger,
		GetResults:                     getResultsFactory(),
		isDebugging:                    isDebugging,
		multipleResultsFromPreviousRun: multipleResultsFromPreviousRun,
		browser:                        b,
	}

	err := s.createTab()
	if err != nil {
		return nil, fmt.Errorf("failed to create tab: %v", err)
	}

	return s, nil
}
