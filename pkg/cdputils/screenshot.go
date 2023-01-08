package cdputils

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
)

// Screenshot makes a screenshot of current window and saves it within a given
// output.
func Screenshot(ctx context.Context, output string) error {
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
		return fmt.Errorf("unable to write file %q: %w", output, err)
	}

	return nil
}
