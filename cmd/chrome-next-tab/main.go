package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io"
	"net/http"
	"os"
)

func main() {
	// -----------------------------------------------------------------------
	// parse args:

	chromeURL := flag.String("url", "http://127.0.0.1:11111", "url")
	flag.Parse()

	// -----------------------------------------------------------------------
	// Get current tab's target id.
	// By default "Target.getTargets" method returns an "unordered" list of target.Info.
	// First or last element doesn't point to current/opened tab, which is sad.

	resp, err := http.Get(fmt.Sprintf("%v/json", *chromeURL))
	if err != nil {
		fmt.Printf("ERROR: Cannot get current tab id: %s\n", err)
		os.Exit(1)
	}

	// read json body:
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR: Cannot get response body: %s\n", err)
		os.Exit(1)
	}

	// map json to object:
	var jsonData []map[string]interface{}
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		fmt.Printf("ERROR: Cannot unmarshal json response: %v\n", err)
		os.Exit(1)
	}

	// choose only page/tabs:
	var jsonDataFiltered []map[string]interface{}
	for _, v := range jsonData {
		if v["type"] == "page" {
			jsonDataFiltered = append(jsonDataFiltered, v)
		}
	}

	// quit if switching not possible:
	jsonDataFilteredLength := len(jsonDataFiltered)

	if jsonDataFilteredLength <= 1 {
		fmt.Printf("WARNING: There are not enough tabs for switching! Quiting ...\n")
		os.Exit(0)
	}

	// calc position of the "next" tab:
	switchToTabTargetID := fmt.Sprintf("%v", jsonDataFiltered[jsonDataFilteredLength-1]["id"])

	// connect to browser:
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), *chromeURL)
	defer cancel()

	ctx, _ := chromedp.NewContext(allocatorCtx)

	// get unordered targets:
	targets, err := chromedp.Targets(ctx)
	if err != nil {
		fmt.Printf("ERROR: Cannot get tab targets: %v\n", err)
		os.Exit(1)
	}

	// bring "next" tab to light:
	for _, v := range targets {
		if string(v.TargetID) == switchToTabTargetID {
			tabCtx, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(v.TargetID))

			if err := chromedp.Run(tabCtx, page.BringToFront()); err != nil {
				fmt.Printf("ERROR: Cannot switch to tab: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("INFO: Tab switched: %v -> %v", jsonDataFiltered[0]["title"], v.Title)
			break
		}
	}
}
