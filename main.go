package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
)

var ctx, cancel = context.WithCancel(context.Background())
var request = make(chan string, 10)
var wg sync.WaitGroup

var windowWidth = 1280
var windowHeight = 800

func crowl(page *agouti.Page, url string) error {
	defer wg.Done()
	if err := page.Navigate(url); err != nil {
		return err
	}
	if err := capture(page, ".", windowWidth, windowHeight); err != nil {
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

func worker(job chan string) error {
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			fmt.Sprintf("--window-size=%d,%d", windowWidth, windowHeight),
			// https://superuser.com/questions/1189725/disable-chrome-scaling
			"--high-dpi-support=1",
			"--force-device-scale-factor=1",
		}),
	)
	if err := driver.Start(); err != nil {
		log.Fatal(err)
	}
	defer driver.Stop()

	page, err := driver.NewPage()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case url := <-job:
			fmt.Println(url)
			if err := crowl(page, url); err != nil {
				panic(err)
			}
		}
	}
}

func capture(page *agouti.Page, destDir string, width int, minheight int) error {
	var pageheight int
	if err := page.RunScript("return document.body.scrollHeight", nil, &pageheight); err != nil {
		return err
	}
	if pageheight > minheight {
		page.Size(width, pageheight)
	} else {
		page.Size(width, minheight)
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
	parallel := 1
	baseurl := "http://localhost:8080/"

	var job = make(chan string, 10)
	go func() {
		var queued []string
		for {
			select {
			case <-ctx.Done():
				return
			case url := <-request:
				if isNewRequest(url, baseurl, queued) {
					if len(queued) > 10 { // limit for test
						break
					}
					queued = append(queued, url)
					wg.Add(1)
					job <- url
				}
			}
		}
	}()

	request <- baseurl
	for i := 0; i < parallel; i++ {
		go worker(job)
	}
	time.Sleep(1 * time.Second)
	wg.Wait()
	cancel()
}
