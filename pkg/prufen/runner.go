package prufen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// TODO: make an iface
type Runner struct {
	baseCtx        context.Context
	opts           []func(*chromedp.ExecAllocator)
	debugf         func(string, ...any)
	runTimeout     time.Duration
	makeScreenshot bool
}

type Options struct {
	// BaseContext is a basic context which is used to pass to the chrome
	// allocator.
	BaseContext context.Context
	// ChromeAllocatorOptions passed to the allocator, some default options
	// always apply.
	ChromeAllocatorOptions []func(*chromedp.ExecAllocator)
	// OperationTimeout is an overall timeout for a single check for an
	// appointment.
	OperationTimeout time.Duration
	// DebugFunc is a function to enable and print debug to an output.
	DebugFunc func(string, ...any)

	// TODO: roundtripper + port ?; baseurl !?; metrics server ?; telegram api; graceful shutdown
	// poll timeout
}

func NewRunner(options Options) *Runner {
	options = setDefaults(options)
	return &Runner{
		baseCtx:    options.BaseContext,
		opts:       options.ChromeAllocatorOptions,
		debugf:     options.DebugFunc,
		runTimeout: options.OperationTimeout,
	}
}

func (r *Runner) Run() (string, error) {
	return r.RunOnce()
}

func (r *Runner) RunOnce() (string, error) {
	ctx, cancel := chromedp.NewExecAllocator(r.baseCtx, r.opts...)
	defer cancel() // allocator

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(r.debugf))
	defer cancel() // new tab

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

	// TODO: extend case + normilize ?? and change xpaths
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
		// TODO: poll? /html/body/div[1], body > div.loading
		// TODO: loading, wait until visible -> not visible
		// TODO: >>>>>>>>>>>>><<<<<<<<<<<<<<<<<<<<<< problem!!!!!!!!!!!!!
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

	// TODO: wait for the list of if success
	var s string
	_ = s
	checkError := []chromedp.Action{chromedp.Sleep(time.Second * 3), chromedp.WaitReady(`//*[@id="messagesBox"]/ul/li`, chromedp.BySearch)}
	_ = checkError

	var summurySteps []chromedp.Action
	for _, stepsSlice := range [][]chromedp.Action{
		preSteps, czSteps, applicantsNumberSteps,
		liveInBerlinSteps, memberCZSteps, postSteps,
		screenshotStep,
	} {
		summurySteps = append(summurySteps, stepsSlice...)
	}

	var u string
	summurySteps = append(summurySteps, chromedp.Location(&u))

	if err := chromedp.Run(ctx, summurySteps...); err != nil {
		return "", fmt.Errorf("failed to run chrome: %w", err)
	}

	return u, nil
}

const (
	DefaultUserAgent    = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`
	DefaultWindowWidth  = 1200
	DefaultWindowHeight = 800

	DefaultContextTimeout = time.Minute * 5
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

	if options.OperationTimeout == 0 {
		options.OperationTimeout = DefaultContextTimeout
	}
	// sanity check
	if options.OperationTimeout < time.Second*30 {
		options.OperationTimeout = DefaultContextTimeout
	}

	if options.BaseContext == nil {
		options.BaseContext = context.Background()
	}

	return options
}
