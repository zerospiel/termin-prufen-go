package prufen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func getOptionsSteps(
	name string,
	optionsSelName string,
	searchValue string,
	awaitingNextSel string,
) []chromedp.Action {
	var nodes []*cdp.Node
	return []chromedp.Action{
		chromedp.Nodes(optionsSelName, &nodes, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
		chromedp.ActionFunc(func(ctx context.Context) error {
			val, ok := findValueAmongNodes(nodes, searchValue)
			if !ok {
				return fmt.Errorf("%s: no value for %q among attributes has been found, len nodes %d", name, searchValue, len(nodes))
			}
			return chromedp.SetValue(optionsSelName, val, chromedp.BySearch).Do(ctx)
		}),
		chromedp.Sleep(time.Millisecond * 250),
		chromedp.WaitVisible(awaitingNextSel, chromedp.BySearch),
		chromedp.Sleep(time.Millisecond * 250),
	}
}

func findValueAmongNodes(nodes []*cdp.Node, value string) (string, bool) {
	value = strings.ToLower(value)
	for _, n := range nodes {
		for _, c := range n.Children {
			val, ok := findValueInAttributes(value, c.Attributes...)
			for _, ccnode := range c.Children {
				if strings.ToLower(ccnode.NodeValue) == value {
					return sliceKV2map(c.Attributes)["value"], true
				}
			}
			if ok {
				return val, ok
			}
		}
	}

	return "", false
}

func findValueInAttributes(search string, kvs ...string) (string, bool) {
	if len(kvs)&1 != 0 { // not even
		return "", false
	}

	m := sliceKV2map(kvs)

	for _, v := range m {
		if strings.ToLower(v) == search {
			return m["value"], true
		}
	}

	return "", false
}

func sliceKV2map(kvs []string) map[string]string {
	m := make(map[string]string, len(kvs)/2)
	for i := 0; i < len(kvs)/2; i++ {
		m[kvs[i*2]] = kvs[i*2+1]
	}
	return m
}
