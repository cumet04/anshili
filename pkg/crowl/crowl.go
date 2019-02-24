package crowl

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sclevine/agouti"
)

func CrowlOne(page *agouti.Page, url string) ([]string, error) {
	var windowWidth = 1280
	var windowHeight = 800

	if err := page.Navigate(url); err != nil {
		return nil, err
	}

	if err := capture(page, "./dest", windowWidth, windowHeight); err != nil {
		return nil, errors.Wrap(err, "capture failed")
	}

	var links []string
	if err := page.RunScript("return Array.from(document.querySelectorAll('a')).map(e => e.href)", nil, &links); err != nil {
		return nil, err
	}
	return links, nil
}

func CheckIsNewRequest(url string, baseurl string, done []string) bool {
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
