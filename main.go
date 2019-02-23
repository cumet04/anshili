package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
)

var request = make(chan string)

var windowWidth = 1280
var windowHeight = 800

func crowl(page *agouti.Page, url string) error {
	if err := page.Navigate(url); err != nil {
		return err
	}

	if err := capture(page, "./dest", windowWidth, windowHeight); err != nil {
		return errors.Wrap(err, "capture failed")
	}

	var links []string
	if err := page.RunScript("return Array.from(document.querySelectorAll('a')).map(e => e.href)", nil, &links); err != nil {
		return err
	}
	for _, l := range links {
		request <- l
	}
	return nil
}

func capture(page *agouti.Page, destDir string, width int, minheight int) error {
	var pageheight int
	if err := page.RunScript("return document.body.scrollHeight", nil, &pageheight); err != nil {
		return err
	}
	if pageheight > minheight {
		if err := page.Size(width, pageheight); err != nil {
			return err
		}
	} else {
		if err := page.Size(width, minheight); err != nil {
			return err
		}
	}

	// make capture file's path from url
	// http://foo.com/bar/ -> foo.com/bar/index.png
	// https://foo.com/bar/baz.html -> foo.com/bar/baz.png
	url, err := page.URL()
	if err != nil {
		return err
	}
	path := strings.NewReplacer("http://", "", "https://", "", ".html", "").Replace(url)
	if strings.HasSuffix(path, "/") {
		path = path + "index"
	}
	path = filepath.Join(destDir, filepath.FromSlash(path+".png"))
	if err := os.MkdirAll(filepath.Dir(path), 0775); err != nil {
		return err
	}
	if err := page.Screenshot(path); err != nil {
		return err
	}

	return nil
}

func isNewRequest(url string, baseurl string, done []string) bool {
	if !strings.HasPrefix(url, baseurl) {
		return false
	}
	if strings.Contains(url, "?") {
		return false
	}
	for _, u := range done {
		if u == url {
			return false
		}
	}
	return true
}

func main() {
	parallel := 4

	pages := make(chan *agouti.Page, parallel)
	for i := 0; i < parallel; i++ {
		// url := "http://127.0.0.1:4444/wd/hub"
		// page, err := remoteDriver(url, "firefox")
		driver, page, err := localChromeDriver()
		if err != nil {
			panic(err)
		}
		defer driver.Stop()
		pages <- page
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	baseurl := "http://localhost:8080/"
	go func() {
		var queued []string
		for {
			select {
			case <-ctx.Done():
				return
			case url := <-request:
				if !isNewRequest(url, baseurl, queued) {
					break
				}
				if len(queued) > 10 { // limit for test
					break
				}
				queued = append(queued, url)
				wg.Add(1)
				go func() {
					fmt.Println(url)
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					case page := <-pages:
						defer func() { pages <- page }()
						if err := crowl(page, url); err != nil {
							log.Fatalln(err)
						}
					}
				}()
			}
		}
	}()
	request <- baseurl

	wg.Wait()
	cancel()
}
