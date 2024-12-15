package browser

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"huurwoning/logger"

	"github.com/chromedp/chromedp"
)

type Browser struct {
	ctx      context.Context
	cancel   context.CancelFunc
	mutex    sync.Mutex
	logger   *logger.Logger
	debug    bool
	tabCount int
	isAlive  bool
}

const maxRetries = 3

func New(debug bool, globalLogger *logger.GlobalLogger) (*Browser, error) {
	logger := globalLogger.Logger("BROWSER")
	b := &Browser{
		logger: logger,
		debug:  debug,
	}
	err := b.createBrowser()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Browser) IsAlive() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.isAlive
}

func (b *Browser) createBrowser() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-extensions", true),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create a new context with logging if debug is enabled
	var ctx context.Context
	var cancel context.CancelFunc
	if b.debug {
		ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(allocCtx)
	}
	b.ctx = ctx
	b.cancel = cancel

	// Ensure the browser is launched
	if err := chromedp.Run(b.ctx); err != nil {
		return err
	}

	b.logger.Info("New browser instance created")
	b.isAlive = true
	return nil
}

func (b *Browser) Close() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.cancel != nil {
		b.cancel()
		b.isAlive = false
		b.logger.Warn("Browser instance closed")
	}
}

func (b *Browser) RecreateIfNeeded() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if !b.isAlive {
		b.logger.Warn("Browser is not alive, recreating...")
		return b.createBrowser()
	}
	return nil
}

func (b *Browser) CreateTab() (context.Context, context.CancelFunc, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.ctx == nil {
		return nil, nil, errors.New("browser not initialized")
	}

	tabCtx, cancel := chromedp.NewContext(b.ctx)
	b.tabCount++
	b.logger.Info(fmt.Sprintf("Tab count updated: %d", b.tabCount))
	return tabCtx, cancel, nil
}

func (b *Browser) DecreaseTabCount() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.tabCount--
}

func (b *Browser) RunInTab(ctx context.Context, actions ...chromedp.Action) error {
	return chromedp.Run(ctx, actions...)
}
