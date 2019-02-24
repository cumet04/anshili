package crowl

import (
	"context"
	"errors"
	"sync"

	"github.com/sclevine/agouti"
)

var localLock sync.Mutex

type Page struct {
	agouti.Page
	driver *agouti.WebDriver
	lock   *sync.Mutex
}

func (p *Page) Destroy() error {
	defer func() {
		if p.driver != nil {
			p.driver.Stop()
		}
		if p.lock != nil {
			p.lock.Unlock()
		}
	}()
	return p.Page.Destroy()
}

func LocalChromeDriver(ctx context.Context) (*Page, error) {
	localLock.Lock()
	select {
	case <-ctx.Done():
		return nil, errors.New("ctx done")
	default:
	}
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			// https://superuser.com/questions/1189725/disable-chrome-scaling
			"--high-dpi-support=1",
			"--force-device-scale-factor=1",
		}),
	)
	if err := driver.Start(); err != nil {
		localLock.Unlock()
		return nil, err
	}
	page, err := driver.NewPage()
	if err != nil {
		localLock.Unlock()
		driver.Stop()
		return nil, err
	}
	return &Page{
		Page:   *page,
		driver: driver,
		lock:   &localLock,
	}, nil
}

func RemoteDriver(ctx context.Context, url, name string) (*Page, error) {
	options := []agouti.Option{agouti.Browser(name)}
	var c chan *agouti.Page
	var err error
	go func() {
		var p *agouti.Page
		p, err = agouti.NewPage(url, options...)
		c <- p
	}()
	select {
	case <-ctx.Done():
		return nil, errors.New("ctx done")
	case p := <-c:
		if err != nil {
			return nil, err
		}
		return &Page{
			Page:   *p,
			driver: nil,
			lock:   nil,
		}, nil
	}
}
