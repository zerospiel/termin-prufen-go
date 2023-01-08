package prufen

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO: make an iface
type Runner struct {
	baseCtx                 context.Context
	debugf                  func(string, ...any)
	port                    string
	screenshotsPath         string
	opts                    []func(*chromedp.ExecAllocator)
	runTimeout              time.Duration
	gracefulShutdownTimeout time.Duration
}

type Options struct {
	// BaseContext is a basic context which is used to pass to the chrome
	// allocator.
	BaseContext context.Context
	// ChromeAllocatorOptions passed to the allocator, some default options
	// always apply.
	ChromeAllocatorOptions []func(*chromedp.ExecAllocator)
	// ScenarioTimeout is an overall timeout for a single check for an
	// appointment.
	ScenarioTimeout time.Duration
	// DebugFunc is a function to enable and print debug to an output.
	DebugFunc func(string, ...any)
	// ScreenshotsPath enables creation of screenshots after the scenario
	// run, and sets the given path in which screenshots will be stored.
	// Each screenshot has a unique name.
	ScreenshotsPath string

	// PollInterval sets the interval between the scenario runs.
	PollInterval time.Duration
	// GracefulShutdownTimeout defines duration for the shutting down the server.
	GracefulShutdownTimeout time.Duration
	// Port defines the HTTP port of the application.
	Port int

	// TelegramAPIToken is used to call API of Telegram™.
	TelegramAPIToken string
	// TelegramChatID defines the ID of the chat in which messages will be send to.
	TelegramChatID string
}

func NewRunner(options Options) *Runner {
	options = setDefaults(options)
	return &Runner{
		baseCtx:                 options.BaseContext,
		debugf:                  options.DebugFunc,
		port:                    strconv.Itoa(options.Port),
		screenshotsPath:         options.ScreenshotsPath,
		opts:                    options.ChromeAllocatorOptions,
		runTimeout:              options.ScenarioTimeout,
		gracefulShutdownTimeout: options.GracefulShutdownTimeout,
	}
}

var ran uint32

func (r *Runner) Run(ctx context.Context) error {
	if old := atomic.SwapUint32(&ran, 1); old != 0 {
		return fmt.Errorf("runner is already running")
	}
	defer atomic.StoreUint32(&ran, 0)

	server := &http.Server{
		Addr:    r.port,
		Handler: r.setupHandler(),
	}

	var err error
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Printf("server started at %q...", server.Addr)

	<-ctx.Done()
	log.Printf("shutting down the server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), r.gracefulShutdownTimeout)
	defer func() {
		cancel()
	}()

	server.SetKeepAlivesEnabled(false)
	if err = server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown: %s", err)
	}

	log.Printf("successfuly shutted down the server")

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return err
}

func (r *Runner) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO:
		// telegram
		// run_once itself + metrics
		// polling with jitter
	}
}

func (r *Runner) setupHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", r.Handler())

	return mux
}

func (r *Runner) RunOnce() (string, bool, error) {
	ctx, cancel := chromedp.NewExecAllocator(r.baseCtx, r.opts...)
	defer cancel() // allocator

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(r.debugf))
	defer cancel() // new tab

	if err := chromedp.Run(ctx); err != nil {
		return "", false, fmt.Errorf("initial run failed: %w", err)
	}

	ctx, cancel = context.WithTimeout(ctx, r.runTimeout)
	defer cancel() // add timeout to avoid hanging

	preSteps := []chromedp.Action{
		// navigate to the basic page...
		chromedp.Navigate("https://otv.verwalt-berlin.de/ams/TerminBuchen?lang=en"),
		// ... wait the very first page...
		chromedp.WaitVisible(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		// ... and click Book Appointment...
		chromedp.Click(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		// ... wait the consent page...
		chromedp.WaitVisible(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		// ... click the checkbox...
		chromedp.Click(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		// ... click Next on the consent page...
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		// ... wait the page with a citizenship (number of redirects)...
		chromedp.WaitVisible(`//*[@id="xi-fs-19"]`, chromedp.BySearch),
	}

	czSteps := getOptionsSteps("select citizenship", `//*[@id="xi-sel-400"]`, "Russian Federation", `//*[@id="xi-sel-422"]`)

	applicantsNumberSteps := getOptionsSteps("select applicants num", `//*[@id="xi-sel-422"]`, "2", `//*[@id="xi-sel-427"]`)

	liveInBerlinSteps := getOptionsSteps("live in berlin", `//*[@id="xi-sel-427"]`, "yes", `//*[@id="xi-sel-428"]`)

	// TODO: if no
	memberCZSteps := getOptionsSteps("select family member citizenship", `//*[@id="xi-sel-428"]`, "Russian Federation", `//*[@id="xi-div-30"]`)

	postSteps := []chromedp.Action{
		// click on apply for a residence permit...
		chromedp.Click(`//*[@id="xi-div-30"]/div[1]`, chromedp.BySearch),
		// ... wait for reasons for residance permit...
		chromedp.WaitVisible(`//*[@id="inner-160-0-1"]`, chromedp.BySearch),
		// ... click economic activity...
		chromedp.Click(`//*[@id="inner-160-0-1"]/div/div[3]`, chromedp.BySearch),
		// ... click blaukarte...
		chromedp.Click(`//*[@id="SERVICEWAHL_EN160-0-1-1-324659"]`, chromedp.BySearch),
		// ... wait until Next button...
		chromedp.WaitVisible(`//*[@id="applicationForm:managedForm"]/div[5]`, chromedp.BySearch),
		// ... click Next button to find termin...
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		// ...wait the result
		chromedp.WaitVisible(`body > div.loading`, chromedp.ByQuery),
		chromedp.WaitNotVisible(`body > div.loading`, chromedp.ByQuery),
	}

	// TODO: ++ screenshot name ? + actionfunc
	var buf []byte
	const curDir = `/Users/morgoev/go/src/github.com/zerospiel/termin-prufen-go`

	screenShotFile := filepath.Join(curDir, "screenshot.jpg")
	screenshotStep := []chromedp.Action{
		chromedp.FullScreenshot(&buf, 90),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return os.WriteFile(screenShotFile, buf, 0o644)
		}),
	}

	var nodes []*cdp.Node
	checkMessagesBoxElem := []chromedp.Action{
		// if the nodes will be empty then we'd found the appointment
		chromedp.Nodes(`//*[@id="messagesBox"]`, &nodes, chromedp.BySearch, chromedp.AtLeast(0)),
	}

	var summurySteps []chromedp.Action
	for _, stepsSlice := range [][]chromedp.Action{
		preSteps, czSteps, applicantsNumberSteps,
		liveInBerlinSteps, memberCZSteps, postSteps,
		checkMessagesBoxElem, screenshotStep,
	} {
		summurySteps = append(summurySteps, stepsSlice...)
	}

	var u string
	summurySteps = append(summurySteps, chromedp.Location(&u))

	if err := chromedp.Run(ctx, summurySteps...); err != nil {
		return "", false, fmt.Errorf("failed to run chrome: %w", err)
	}

	return u, len(nodes) == 0, nil
}

const (
	DefaultUserAgent    = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`
	DefaultWindowWidth  = 1200
	DefaultWindowHeight = 800

	DefaultPollInterval            = time.Minute * 3
	DefaultScenarioTimeout         = time.Minute * 5
	DefaultGracefulShutdownTimeout = time.Second * 15
	DefaultHTTPPort                = 80
)

func setDefaults(options Options) Options {
	requiredOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.IgnoreCertErrors,
		// any user agent just in case
		chromedp.UserAgent(DefaultUserAgent),
		chromedp.WindowSize(DefaultWindowWidth, DefaultWindowHeight),
		// bypass automation detection
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	if options.ChromeAllocatorOptions == nil {
		options.ChromeAllocatorOptions = requiredOptions
	} else {
		options.ChromeAllocatorOptions = append(requiredOptions, options.ChromeAllocatorOptions...)
	}

	if options.ScenarioTimeout == 0 ||
		options.ScenarioTimeout < time.Second*30 {
		options.ScenarioTimeout = DefaultScenarioTimeout
	}

	if options.BaseContext == nil {
		options.BaseContext = context.Background()
	}

	if options.GracefulShutdownTimeout == 0 {
		options.GracefulShutdownTimeout = DefaultGracefulShutdownTimeout
	}

	if options.Port == 0 {
		options.Port = DefaultHTTPPort
	}

	if options.PollInterval == 0 {
		options.PollInterval = DefaultPollInterval
	}

	return options
}
