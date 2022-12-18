package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	chromeOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.IgnoreCertErrors,
		// any user agent just in case
		chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`),
		chromedp.WindowSize(1200, 800),
		// bypass automation detection
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOptions...)
	defer cancel() // allocator

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))

	defer cancel() // new tab

	ctx, cancel = context.WithTimeout(ctx, time.Minute*5)
	defer cancel() // add timeout 5 mins just to avoid hanging

	// curDir, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	const curDir = `/Users/morgoev/go/src/github.com/zerospiel/termin-bucher-go`

	screenShotFile := filepath.Join(curDir, "screenshot.jpg")
	saveScreenShot := chromedp.ActionFunc(makeScreenshot(ctx, screenShotFile))
	_ = saveScreenShot

	var buf []byte
	var citizenshipNodes []*cdp.Node
	_ = citizenshipNodes

	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://otv.verwalt-berlin.de/ams/TerminBuchen?lang=en"),
		// wait the very first page
		chromedp.WaitVisible(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		chromedp.Click(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
		//
		// wait the consent page
		chromedp.WaitVisible(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		chromedp.Click(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
		//
		// click Next after the consent page
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		// wait the page with a citizenship (number of redirects)
		chromedp.WaitVisible(`//*[@id="xi-fs-19"]`, chromedp.BySearch),
		// select Citizenship
		// chromedp.ActionFunc(func(ctx context.Context) error {
		// 	node, rerr := dom.GetDocument().WithDepth(1).Do(ctx)
		// 	if rerr != nil {
		// 		return rerr
		// 	}
		// 	q.Q(node) // DEBUG
		// 	return nil
		// }),
		// chromedp.Nodes(`//*[@id="xi-sel-400"]`, &citizenshipNodes, chromedp.BySearch),
		// chromedp.SetJavascriptAttribute(`//*[@id="xi-sel-400"]`, "", "Russian Federation", chromedp.BySearch),
		chromedp.SetValue(`//*[@id="xi-sel-400"]`, "160", chromedp.BySearch),
		chromedp.WaitVisible(`//*[@id="xi-sel-422"]`, chromedp.BySearch),
		// TODO: uncommenct, kinda works
		//
		// select Number of applicants
		chromedp.SetValue(`//*[@id="xi-sel-422"]`, "2", chromedp.BySearch),
		chromedp.WaitVisible(`//*[@id="xi-sel-427"]`, chromedp.BySearch),
		//
		// select Do you live in Berlin
		chromedp.SetValue(`//*[@id="xi-sel-427"]`, "1", chromedp.BySearch),
		chromedp.WaitVisible(`//*[@id="xi-sel-428"]`, chromedp.BySearch),
		//
		// select Citizenship of the family member
		chromedp.SetValue(`//*[@id="xi-sel-428"]`, "160-0", chromedp.BySearch),
		//
		// wait for options
		chromedp.WaitVisible(`//*[@id="xi-div-30"]`, chromedp.BySearch),
		// click on apply for a residence permit
		chromedp.Click(`//*[@id="xi-div-30"]/div[1]`, chromedp.BySearch),
		// wait for reasons for residance permit
		chromedp.WaitVisible(`//*[@id="inner-160-0-1"]`, chromedp.BySearch),
		// click economic activity
		chromedp.Click(`//*[@id="inner-160-0-1"]/div/div[3]`, chromedp.BySearch),
		// click blaukarte
		chromedp.Click(`//*[@id="SERVICEWAHL_EN160-0-1-1-324659"]`, chromedp.BySearch),
		// wait until Next button
		chromedp.WaitVisible(`//*[@id="applicationForm:managedForm"]/div[5]`, chromedp.BySearch),
		// click Next button to find termin
		chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
		// wait the result
		chromedp.WaitVisible(`/html/body/div[1]`, chromedp.BySearch),
		chromedp.WaitNotVisible(`/html/body/div[1]`, chromedp.BySearch),
		chromedp.FullScreenshot(&buf, 90),
	); err != nil {
		panic(err)
	}

	if err := os.WriteFile(screenShotFile, buf, 0o644); err != nil {
		panic(err)
	}

	// fmt.Println(">>> first screenshot saved")

	// var nodeID cdp.NodeID
	// nodeID, ok := findValueAmongNodes(citizenshipNodes, "Russian Federation")
	// if !ok {
	// 	panic("not exist")
	// }
	// fmt.Println(">>>>>>", nodeID)
	// if err := chromedp.Run(ctx,
	// 	chromedp.Click([]cdp.NodeID{nodeID}, chromedp.ByNodeID),
	// 	chromedp.FullScreenshot(&buf, 90),
	// ); err != nil {
	// 	panic(err)
	// }

	// if err := os.WriteFile(screenShotFile, buf, 0o644); err != nil {
	// 	panic(err)
	// }
}

func findValueAmongNodes(nodes []*cdp.Node, value string) (cdp.NodeID, bool) {
	for _, n := range nodes {
		for _, c := range n.Children {
			for _, cc := range c.Children {
				if cc.NodeValue == value {
					return c.NodeID, true
				}
			}
		}
	}

	return 0, false
}

func makeScreenshot(ctx context.Context, output string) func(context.Context) error {
	return func(ctx context.Context) error {
		return screenshot(ctx, output)
	}
}

func screenshot(ctx context.Context, output string) error {
	// get layout metrics
	_, _, _, _, _, cssContentSize, err := page.GetLayoutMetrics().Do(ctx)
	if err != nil {
		return fmt.Errorf("unable to get layout metrics: %w", err)
	}

	width, height := int64(math.Ceil(cssContentSize.Width)), int64(math.Ceil(cssContentSize.Height))

	// force viewport emulation
	err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
		WithScreenOrientation(&emulation.ScreenOrientation{
			Type:  emulation.OrientationTypePortraitPrimary,
			Angle: 0,
		}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("unable to override viewport emulation: %w", err)
	}

	// capture screenshot
	buf, err := page.CaptureScreenshot().
		WithClip(&page.Viewport{
			X:      cssContentSize.X,
			Y:      cssContentSize.Y,
			Width:  cssContentSize.Width,
			Height: cssContentSize.Height,
			Scale:  1,
		}).Do(ctx)
	if err != nil {
		return fmt.Errorf("unable to capture a screenshot: %w", err)
	}

	if err := os.WriteFile(output, buf, 0o644); err != nil {
		return fmt.Errorf("unable to write file '%s': %w", output, err)
	}

	return nil
}
