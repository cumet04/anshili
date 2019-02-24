package crowl

import (
	"github.com/sclevine/agouti"
)

func LocalChromeDriver() (*agouti.WebDriver, *agouti.Page, error) {
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			// fmt.Sprintf("--window-size=%d,%d", windowWidth, windowHeight),
			// https://superuser.com/questions/1189725/disable-chrome-scaling
			"--high-dpi-support=1",
			"--force-device-scale-factor=1",
		}),
	)
	if err := driver.Start(); err != nil {
		return nil, nil, err
	}
	page, err := driver.NewPage()
	if err != nil {
		defer driver.Stop()
		return nil, nil, err
	}
	return driver, page, nil
}

func localFirefoxDriver() (*agouti.WebDriver, *agouti.Page, error) {
	driver := agouti.Selenium()
	if err := driver.Start(); err != nil {
		return nil, nil, err
	}
	page, err := driver.NewPage(agouti.Browser("firefox"))
	if err != nil {
		defer driver.Stop()
		return nil, nil, err
	}
	return driver, page, nil
}

func remoteDriver(url, name string) (*agouti.Page, error) {
	options := []agouti.Option{agouti.Browser(name)}
	page, err := agouti.NewPage(url, options...)
	if err != nil {
		return nil, err
	}
	return page, nil
}
