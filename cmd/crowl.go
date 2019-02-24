package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cumet04/anshili/pkg/crowl"
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
					defer wg.Done()
					page, err := crowl.LocalChromeDriver(ctx)
					if err != nil {
						log.Fatalln(err)
					}
					defer page.Destroy()
					fmt.Println(url)
					links, err := crowl.CrowlOne(&page.Page, url)
					for _, l := range links {
						request <- l
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
