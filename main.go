package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/zerospiel/termin-prufen-go/pkg/prufen"
)

func main() {
	runner := prufen.NewRunner(prufen.Options{
		DebugFunc: log.Printf,
	})
	fmt.Println(runner.RunOnce())
	// chromeOptions := append(chromedp.DefaultExecAllocatorOptions[:],
	// 	chromedp.NoSandbox,
	// 	chromedp.IgnoreCertErrors,
	// 	// any user agent just in case
	// 	chromedp.UserAgent(`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`),
	// 	chromedp.WindowSize(1200, 800),
	// 	// bypass automation detection
	// 	chromedp.Flag("enable-automation", false),
	// 	chromedp.Flag("disable-blink-features", "AutomationControlled"),
	// )

	// ctx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOptions...)
	// defer cancel() // allocator

	// // ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	// ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(nil))

	// defer cancel() // new tab

	// ctx, cancel = context.WithTimeout(ctx, time.Minute*5)
	// defer cancel() // add timeout 5 mins just to avoid hanging

	// // curDir, err := os.Executable()
	// // if err != nil {
	// // 	panic(err)
	// // }
	// const curDir = `/Users/morgoev/go/src/github.com/zerospiel/termin-prufen-go`

	// screenShotFile := filepath.Join(curDir, "screenshot.jpg")
	// saveScreenShot := chromedp.ActionFunc(makeScreenshot(ctx, screenShotFile))
	// _ = saveScreenShot

	// var buf []byte
	// var citizenshipNodes []*cdp.Node
	// _ = citizenshipNodes

	// if err := chromedp.Run(ctx,
	// 	chromedp.Navigate("https://otv.verwalt-berlin.de/ams/TerminBuchen?lang=en"),
	// 	// wait the very first page
	// 	chromedp.WaitVisible(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
	// 	chromedp.Click(`//*[@id="mainForm"]/div/div/div/div/div/div/div/div/div/div[1]/div[1]/div[2]/a`, chromedp.BySearch),
	// 	//
	// 	// wait the consent page
	// 	chromedp.WaitVisible(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
	// 	chromedp.Click(`//*[@id="xi-cb-1"]`, chromedp.BySearch),
	// 	//
	// 	// click Next after the consent page
	// 	chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
	// 	// wait the page with a citizenship (number of redirects)
	// 	chromedp.WaitVisible(`//*[@id="xi-fs-19"]`, chromedp.BySearch),
	// 	// select Citizenship
	// 	// chromedp.ActionFunc(func(ctx context.Context) error {
	// 	// 	node, rerr := dom.GetDocument().WithDepth(1).Do(ctx)
	// 	// 	if rerr != nil {
	// 	// 		return rerr
	// 	// 	}
	// 	// 	q.Q(node) // DEBUG
	// 	// 	return nil
	// 	// }),
	// 	chromedp.Nodes(`//*[@id="xi-sel-400"]`, &citizenshipNodes, chromedp.BySearch),
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		q.Q(citizenshipNodes)
	// 		val, ok := findValueAmongNodes(citizenshipNodes, "Russian Federation")
	// 		if !ok {
	// 			return fmt.Errorf("no value, len nodes %d", len(citizenshipNodes))
	// 		}
	// 		return chromedp.SetValue(`//*[@id="xi-sel-400"]`, val, chromedp.BySearch).Do(ctx)
	// 	}),
	// 	// chromedp.SetValue(`//*[@id="xi-sel-400"]`, "160", chromedp.BySearch),
	// 	// chromedp.WaitVisible(`//*[@id="xi-sel-422"]`, chromedp.BySearch),
	// 	// TODO: uncommenct, kinda works
	// 	// //
	// 	// // select Number of applicants
	// 	// chromedp.SetValue(`//*[@id="xi-sel-422"]`, "2", chromedp.BySearch),
	// 	// chromedp.WaitVisible(`//*[@id="xi-sel-427"]`, chromedp.BySearch),
	// 	// //
	// 	// // select Do you live in Berlin
	// 	// chromedp.SetValue(`//*[@id="xi-sel-427"]`, "1", chromedp.BySearch),
	// 	// chromedp.WaitVisible(`//*[@id="xi-sel-428"]`, chromedp.BySearch),
	// 	// //
	// 	// // select Citizenship of the family member
	// 	// chromedp.SetValue(`//*[@id="xi-sel-428"]`, "160-0", chromedp.BySearch),
	// 	// //
	// 	// // wait for options
	// 	// chromedp.WaitVisible(`//*[@id="xi-div-30"]`, chromedp.BySearch),
	// 	// // click on apply for a residence permit
	// 	// chromedp.Click(`//*[@id="xi-div-30"]/div[1]`, chromedp.BySearch),
	// 	// // wait for reasons for residance permit
	// 	// chromedp.WaitVisible(`//*[@id="inner-160-0-1"]`, chromedp.BySearch),
	// 	// // click economic activity
	// 	// chromedp.Click(`//*[@id="inner-160-0-1"]/div/div[3]`, chromedp.BySearch),
	// 	// // click blaukarte
	// 	// chromedp.Click(`//*[@id="SERVICEWAHL_EN160-0-1-1-324659"]`, chromedp.BySearch),
	// 	// // wait until Next button
	// 	// chromedp.WaitVisible(`//*[@id="applicationForm:managedForm"]/div[5]`, chromedp.BySearch),
	// 	// // click Next button to find termin
	// 	// chromedp.Click(`//*[@id="applicationForm:managedForm:proceed"]`, chromedp.BySearch),
	// 	// // wait the result
	// 	// // TODO: poll?
	// 	// chromedp.WaitVisible(`/html/body/div[1]`, chromedp.BySearch),
	// 	// chromedp.WaitNotVisible(`/html/body/div[1]`, chromedp.BySearch),
	// 	chromedp.FullScreenshot(&buf, 90),
	// ); err != nil {
	// 	panic(err)
	// }

	// if err := os.WriteFile(screenShotFile, buf, 0o644); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(">>> first screenshot saved")

	// nodeID, ok := findValueAmongNodes(citizenshipNodes, "Russian Federation")
	// if !ok {
	// 	// panic("not exist")
	// }
	// fmt.Println(">>>>>>", nodeID)
	// /*
	// 	#xi-sel-400 > option:nth-child(136)
	// 	document.querySelector("#xi-sel-400 > option:nth-child(136)")
	// 	//*[@id="xi-sel-400"]/option[136]
	// 	/html/body/div[2]/div[2]/div[4]/div[2]/form/div[2]/div/div[2]/div[8]/div[2]/div[2]/div[1]/fieldset/div[1]/select/option[136]
	// */
	// var u string
	// if err := chromedp.Run(ctx,
	// 	// chromedp.Click(nodeID, chromedp.BySearch),
	// 	// chromedp.SetValue(`//*[@id="xi-sel-400"]`, nodeID, chromedp.BySearch),
	// 	// chromedp.Node(nil),
	// 	// chromedp.Click([]cdp.NodeID{nodeID[0]}, chromedp.ByNodeID),
	// 	// chromedp.Click([]cdp.NodeID{nodeID[1]}, chromedp.ByNodeID),
	// 	chromedp.FullScreenshot(&buf, 90),
	// 	chromedp.Location(&u),
	// ); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(u)

	// if err := os.WriteFile(screenShotFile, buf, 0o644); err != nil {
	// 	panic(err)
	// }
}

func findSelectorIdInKVs(search string, kvs ...string) (string, bool) {
	if len(kvs)&1 != 0 { // not even
		return "", false
	}

	// no sure if kvs would always contain id at the end
	// that's why no searching in the array

	m := make(map[string]string, len(kvs)/2)
	for i := 0; i < len(kvs)/2; i++ {
		m[kvs[i*2]] = kvs[i*2+1]
	}
	for _, v := range m {
		if v == search {
			return m["value"], true
		}
	}

	return "", false
}

func findValueAmongNodes(nodes []*cdp.Node, value string) (string, bool) {
	for _, n := range nodes {
		for _, c := range n.Children {
			// fmt.Printf("c.AttributeValue(\"col0\"): %v\n", c.AttributeValue("col0"))
			_ = value
			sel, ok := findSelectorIdInKVs(value, c.Attributes...)
			if ok {
				return sel, ok
			}
			// q.Q(findSelectorIdInKVs(value, c.Attributes...))
			// for _, cc := range c.Children {
			// 	if cc.NodeValue == value {
			// 		fmt.Printf("n.LocalName: %v\n", n.LocalName)
			// 		fmt.Printf("n.Name: %v\n", n.Name)
			// 		fmt.Printf("n.NodeName: %v\n", n.NodeName)
			// 		fmt.Printf("cc.LocalName: %v\n", cc.LocalName)
			// 		fmt.Printf("cc.Name: %v\n", cc.Name)
			// 		fmt.Printf("cc.NodeName: %v\n", cc.NodeName)
			// 		fmt.Printf("cc.Value: %v\n", cc.Value)
			// 		fmt.Printf("cc.ChildNodeCount: %v\n", cc.ChildNodeCount)
			// 		return []cdp.NodeID{n.NodeID, cc.NodeID}, true
			// 	}
			// }
		}
	}

	return "", false
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
