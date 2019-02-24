package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cumet04/anshili/pkg/crowl"
	"github.com/sclevine/agouti"
	"github.com/spf13/cobra"
)

var crowlCmd = &cobra.Command{
	Use:   "crowl",
	Short: "Crowl a web site and capture each page",
	RunE: func(cmd *cobra.Command, args []string) error {
		return crowlFunc()
	},
}

func crowlFunc() error {
	parallel := 4

	pages := make(chan *agouti.Page, parallel)
	for i := 0; i < parallel; i++ {
		// url := "http://127.0.0.1:4444/wd/hub"
		// page, err := remoteDriver(url, "firefox")

		driver, page, err := crowl.LocalChromeDriver()
		if err != nil {
			panic(err)
		}
		defer driver.Stop()
		pages <- page
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	baseurl := "http://localhost:8080/"
	request := make(chan string)
	go func() {
		var queued []string
		for {
			select {
			case <-ctx.Done():
				return
			case url := <-request:
				if !crowl.CheckIsNewRequest(url, baseurl, queued) {
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
						links, err := crowl.CrowlOne(page, url)
						if err != nil {
							log.Fatalln(err)
						}
						for _, l := range links {
							request <- l
						}
					}
				}()
			}
		}
	}()
	request <- baseurl

	wg.Wait()
	cancel()
	return nil
}
