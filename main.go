package main

import (
	"log"

	"github.com/sclevine/agouti"
)

func main() {
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=1280,800",
			// https://superuser.com/questions/1189725/disable-chrome-scaling
			"--high-dpi-support=1",
			"--force-device-scale-factor=1",
		}),
		agouti.Debug,
	)
	if err := driver.Start(); err != nil {
		log.Fatal(err)
	}
	defer driver.Stop()
	page, err := driver.NewPage()
	if err != nil {
		log.Fatal(err)
	}
	page.Navigate("https://godoc.org/github.com/sclevine/agouti")
	var height int
	page.RunScript("return document.body.scrollHeight", nil, &height)
	if height > 800 {
		page.Size(1280, height)
	} else {
		page.Size(1280, 800)
	}
	page.Screenshot("capture.png")
}
