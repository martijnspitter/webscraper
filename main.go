package main

import (
	"log"
	"time"

	"huurwoning/beumer"
	"huurwoning/bouwinvest"
	"huurwoning/browser"
	"huurwoning/config"
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

	logger := globalLogger.Logger("MAIN")

	b, err := browser.New(config.DEBUG_MODE, globalLogger)
	if err != nil {
		log.Fatalf("Failed to create browser: %v", err)
	}
	defer b.Close()

	reboPrevResults := make(map[string]struct{})
	// Initialize reboPrevResults with the provided values
	for _, address := range []string{
		"Operettelaan 717", "Lange Nieuwstraat 151", "W.F. Hermansstraat 268", "Vlechtdraadhof 43", "Huis te Zuylenlaan 43", "Operettelaan 699", "Hof van Oslo 104", "Waterstede 24", "Luzernevlinder 34", "Anna Blamanstraat 2-7", "Amerikalaan 55", "Bisschop Tulpijndreef 55", "Lange Nieuwstraat 183", "Stinzenlaan Noord 165",
	} {
		reboPrevResults[address] = struct{}{}
	}

	vestedaPrevResults := make(map[string]struct{})
	// Initialize vestedaPrevResults with the provided values
	for _, address := range []string{
		"Secretaris Versteeglaan 10", "Wenenpromenade 102", "Waterstraat 30", "Churchilllaan 227", "J. Homan van der Heideplein 10", "Hof van Bern 7", "Kobaltpad 13", "Broederschaplaan 16",
	} {
		vestedaPrevResults[address] = struct{}{}
	}

	bouwInvestPrevResults := make(map[string]struct{})
	// Initialize bouwInvestPrevResults with the provided values
	for _, address := range []string{
		"Haarzicht", "Zijdebalen blok IV", "Rachmaninoffplantsoen 176", "Batau Noord III", "Tuinpark I", "Parkwijk Noord", "Langerak I", "Rachmaninoffhuis", "Zijdebalenstraat 73", "Veemarkt I", "Terwijde 14/15 I", "Galecop I", "Langerak II", "Melissekade 30", "Lloyd Webberhof 21", "Prinses Marijkelaan 135", "De Bongerd I", "Terwijde Zuid", "Vredenburgplein", "Parkwijk Zuid veld 22", "Langerak III", "Prinses Marijkelaan 225", "Zijdebalen III", "Meyster's Buiten I", "Toos Korvezeepad 11", "Licht en Lucht", "Zijdebalen II", "Rolderdiephof 109", "Melissekade 259", "Veemarkt City", "De Bongerd II", "Terwijde 14/15 II", "Parkwijk Het Zand", "Tuinpark II", "Houtrakgracht 326", "Zijdebalen I", "Meyster's Buiten II", "Parkwijk Zuid veld 25", "Van der Marckhof", "Veemarkt Portiek", "Galecop II", "Dichterswijk", "Legends",
	} {
		bouwInvestPrevResults[address] = struct{}{}
	}

	beumerPrevResults := make(map[string]struct{})
	for _, address := range []string{
		"Erroll Garnerstraat 51", "Keulsekade 132 A", "Keulsekade 131 A", "Sonny Rollinsstraat 270", "Erroll Garnerstraat 107", "Erroll Garnerstraat 69",
	} {
		beumerPrevResults[address] = struct{}{}
	}

	reboMultipleResultsFromPreviousRun := &[]string{}
	vestedaMultipleResultsFromPreviousRun := &[]string{}
	bouwInvestMultipleResultsFromPreviousRun := &[]string{}
	beumerPrevResultsFromPreviousRun := &[]string{}

	// Start separate timers for each scraping process
	for {
		err := rebo.Rebo(b, globalLogger, "https://rebowonenhuur.nl/login", reboPrevResults, reboMultipleResultsFromPreviousRun)
		if err != nil {
			logger.Error("Error in Rebo scraping!", "error", err)
		}

		err = vesteda.Vesteda(b, globalLogger, "https://hurenbij.vesteda.com/login", vestedaPrevResults, vestedaMultipleResultsFromPreviousRun)
		if err != nil {
			logger.Error("Error in Vesteda scraping!", "error", err)
		}

		err = bouwinvest.BouwInvest(b, globalLogger, "https://www.wonenbijbouwinvest.nl/huuraanbod?query=Utrecht&range=10&seniorservice=false&order=recent&size=50", bouwInvestPrevResults, bouwInvestMultipleResultsFromPreviousRun)
		if err != nil {
			logger.Error("Error in BouwInvest scraping!", "error", err)
		}

		err = beumer.Beumer(b, globalLogger, "https://www.beumer.nl/huurwoningen/?search=Utrecht&status%5B0%5D=te-huur", beumerPrevResults, beumerPrevResultsFromPreviousRun)
		if err != nil {
			logger.Error("Error in Beumer scraping!", "error", err)
		}

		time.Sleep(30 * time.Second)
	}
}
