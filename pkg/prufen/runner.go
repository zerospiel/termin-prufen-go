package prufen

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

type Runner struct {
	logger *slog.Logger

	botClient *tgbotapi.BotAPI

	baseCtx         context.Context
	debugf          func(string, ...any)
	port            string
	screenshotsPath string

	citizenship             string
	peopleNumber            string
	liveInBerlin            string
	familyMemberCitizenship string
	reason                  string

	telegramAPIToken string
	telegramChatID   int64

	opts                    []func(*chromedp.ExecAllocator)
	runTimeout              time.Duration
	pollInterval            time.Duration
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
	TelegramChatID int64

	// TODO: comment
	Citizenship             string
	PeopleNumber            string
	LiveInBerlin            string
	FamilyMemberCitizenship string
	Reason                  string

	Logger *slog.Logger
}

func NewRunner(options Options) (*Runner, error) {
	options = setDefaults(options)

	r := &Runner{
		logger: options.Logger,

		baseCtx:         options.BaseContext,
		debugf:          options.DebugFunc,
		port:            strconv.Itoa(options.Port),
		screenshotsPath: options.ScreenshotsPath,

		opts:                    options.ChromeAllocatorOptions,
		runTimeout:              options.ScenarioTimeout,
		pollInterval:            options.PollInterval,
		gracefulShutdownTimeout: options.GracefulShutdownTimeout,

		citizenship:             options.Citizenship,
		peopleNumber:            options.PeopleNumber,
		liveInBerlin:            options.LiveInBerlin,
		familyMemberCitizenship: options.FamilyMemberCitizenship,
		reason:                  options.Reason,

		telegramAPIToken: options.TelegramAPIToken,
		telegramChatID:   options.TelegramChatID,
	}

	api, err := tgbotapi.NewBotAPI(r.telegramAPIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to construct telegram bot API: %w", err)
	}
	r.botClient = api

	return r, nil
}

var ran uint32

// Run starts the server with runnable poller and metric for debug purposes.
func (r *Runner) Run(ctx context.Context) error {
	if old := atomic.SwapUint32(&ran, 1); old != 0 {
		return fmt.Errorf("runner is already running")
	}
	defer atomic.StoreUint32(&ran, 0)

	server := &http.Server{
		Addr:    ":" + r.port,
		Handler: r.setupMetricsHandler(),
	}

	var err error
	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.logger.Error("listen failed", "error", err)
			panic(err)
		}
	}()
	r.logger.InfoCtx(ctx, "server started...", "addr", server.Addr)

	go r.poll(ctx)

	<-ctx.Done()
	r.logger.Warn("shutting down the server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), r.gracefulShutdownTimeout)
	defer func() {
		cancel()
	}()

	server.SetKeepAlivesEnabled(false)
	if err = server.Shutdown(ctxShutDown); err != nil {
		r.logger.Error("server Shutdown failed", "error", err)
		return err
	}

	r.logger.Info("successfuly shutted down the server")

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return err
}

func (r *Runner) poll(ctx context.Context) {
	timer := time.NewTicker(r.pollInterval)

	r.RunFullCycle()

	for {
		select {
		case <-timer.C:
			r.RunFullCycle()

		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

// SendMessage sends a given payload to the Telegram chat.
func (r *Runner) SendMessage(payload string) error {
	msgcfg := tgbotapi.NewMessage(r.telegramChatID, payload)
	msgcfg.AllowSendingWithoutReply = true
	msgcfg.DisableWebPagePreview = true

	_, err := r.botClient.Send(msgcfg)
	if err != nil {
		return fmt.Errorf("failed to send message to telegram chat: %w", err)
	}

	return nil
}

// RunOnce runs the full cycle through the ABH/LEA site and
// returns the URI to continue booking the appointment.
func (r *Runner) RunOnce() (uri string, found bool, _ error) {
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
		chromedp.Sleep(time.Millisecond * 250),
		// ... wait the very first page...
		chromedp.WaitVisible(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... and click Book Appointment...
		chromedp.Click(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... wait the consent page...
		chromedp.WaitVisible(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... click the checkbox...
		chromedp.Click(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... click Next on the consent page...
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... wait the page with a citizenship (number of redirects)...
		chromedp.WaitVisible(`//*[@id="xi-fs-19"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
	}

	czSteps := getOptionsSteps("select citizenship", `//*[@id="xi-sel-400"]`, r.citizenship, `//*[@id="xi-sel-422"]`)

	applicantsNumberSteps := getOptionsSteps("select applicants num", `//*[@id="xi-sel-422"]`, r.peopleNumber, `//*[@id="xi-sel-427"]`)

	liveInBerlinSteps := getOptionsSteps("live in berlin", `//*[@id="xi-sel-427"]`, r.liveInBerlin, `//*[@id="xi-sel-428"]`)

	var memberCZSteps []chromedp.Action
	if r.familyMemberCitizenship != "" {
		memberCZSteps = getOptionsSteps("select family member citizenship", `//*[@id="xi-sel-428"]`, r.familyMemberCitizenship, `//*[@id="xi-div-30"]`)
	}

	postSteps := []chromedp.Action{
		// click on apply for a residence permit...
		// TODO: apply or extend
		chromedp.Click(`//*[@id="xi-div-30"]/div[1]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... wait for reasons for residance permit...
		chromedp.WaitVisible(`//*[@id="inner-160-0-1"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... click economic activity...
		chromedp.Click(`//*[@id="inner-160-0-1"]/div/div[3]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... click blaukarte...
		chromedp.Click(`//*[@id="SERVICEWAHL_EN160-0-1-1-324659"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... wait until Next button...
		chromedp.WaitVisible(`//*[@id="applicationForm:managedForm"]/div[5]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ... click Next button to find termin...
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		// ...wait the result
		chromedp.WaitVisible(`body > div.loading`, chromedp.ByQuery),
		chromedp.WaitNotVisible(`body > div.loading`, chromedp.ByQuery),
	}

	var screenshotStep []chromedp.Action
	if r.screenshotsPath != "" {
		var buf []byte

		screenShotFile := filepath.Join(r.screenshotsPath, fmt.Sprintf("screenshot_at_%s.jpg", time.Now().Format(time.DateTime)))

		screenshotStep = []chromedp.Action{
			chromedp.FullScreenshot(&buf, 90),
			chromedp.ActionFunc(func(ctx context.Context) error {
				return os.WriteFile(screenShotFile, buf, 0o644)
			}),
		}
	}

	var nodes []*cdp.Node
	checkMessagesBoxElem := []chromedp.Action{
		// if the nodes will be empty then we'd found the appointment
		chromedp.Nodes(`//*[@id="messagesBox"]`, &nodes, chromedp.BySearch, chromedp.AtLeast(0)),
	}

	var summurySteps []chromedp.Action
	for _, stepsSlice := range [][]chromedp.Action{
		preSteps,
		czSteps,
		applicantsNumberSteps,
		liveInBerlinSteps,
		memberCZSteps,
		postSteps,
		checkMessagesBoxElem,
		screenshotStep,
	} {
		if len(stepsSlice) > 0 {
			summurySteps = append(summurySteps, stepsSlice...)
		}
	}

	var u string
	summurySteps = append(summurySteps, chromedp.Location(&u))

	if err := chromedp.Run(ctx, summurySteps...); err != nil {
		return "", false, fmt.Errorf("failed to run chrome: %w", err)
	}

	return u, len(nodes) == 0, nil
}

// RunFullCycle is used mostly as one-liner, it consists of
// running the Runner.RunOnce() and
// the most simple retries while sending message to Telegram.
func (r *Runner) RunFullCycle() {
	now := time.Now()
	r.logger.Debug("new poll cycle")
	continueURI, successfull, err := r.RunOnce()
	if err != nil {
		r.logger.Error("failed to check", "error", err)
		return
	}
	r.logger.Debug("fetched one run", "elapsed sec", time.Since(now).Seconds())

	scenariosTotal.Inc()
	if successfull {
		successScenariosTotal.Inc()
	}

	if !successfull && r.debugf == nil {
		r.logger.Info("checked, no available slots")
		return
	}

	text := "Slots are available!\n"
	if !successfull {
		text = "No slots are available\n"
	}
	text += fmt.Sprintf("Proceed further: %s", continueURI)

	// 1sec retry code block
	for i := 0; i < 5; i++ {
		if err := r.SendMessage(text); err != nil {
			r.logger.Error("failed to send message to telegram", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	r.logger.Debug("poll ended")
}

func (r *Runner) setupMetricsHandler() http.Handler {
	mux := http.NewServeMux()
	http.NewServeMux().Handle("/metrics", promhttp.Handler())

	return mux
}

const (
	DefaultUserAgent    = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`
	DefaultWindowWidth  = 1200
	DefaultWindowHeight = 800

	DefaultPollInterval            = time.Minute * 3
	DefaultScenarioTimeout         = time.Second * 50
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

	if options.Logger == nil {
		options.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return options
}
