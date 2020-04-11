package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

func HttpMatch(ctx *cli.Context) error {

	uriStr := ctx.Args().First()
	_, err := url.ParseRequestURI(uriStr)
	if err != nil {
		log.Errorln("Please enter a correct URI.")
		cli.ShowCommandHelpAndExit(ctx, ctx.Command.Name, 1)
	}
	expect := ctx.Args().Get(1)
	if expect == "" {
		log.Errorln("expect cann't be empty.")
		cli.ShowCommandHelpAndExit(ctx, ctx.Command.Name, 1)
	}
	times := ctx.Int("times")
	interval := ctx.Int("interval")
	timeout := ctx.Int("timeout")
	timeoutChan := make(chan bool, 1)
	log.Debugln("times", times, "timeout", timeout)
	go func() {
		for i := 0; i < times; i++ {
			resp, err := http.Get(uriStr)
			if err != nil {
				log.Warnln(err)
			}
			if resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Warnln(err)
				}
				bodyString := string(bodyBytes)
				log.Debugln("body", bodyString)
				matched, _ := regexp.MatchString(expect, bodyString)
				log.Debugln("matched", matched)
				if matched {
					timeoutChan <- true
					break
				}
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
		timeoutChan <- false
	}()
	select {
	case res := <-timeoutChan:
		if res {
			return nil
		} else {
			log.Debugln("Not Match")
		}
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Debugln("Timeout")
	}
	return fmt.Errorf("[%s] not in response body.", expect)
}

func main() {
	HttpMatch := cli.Command{
		Name:      "match",
		Usage:     "match expect content for response body",
		ArgsUsage: "[URI] [expect]",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "times", Value: 10, Usage: "how many requests to make"},
			&cli.IntFlag{Name: "interval", Value: 3, Usage: "interval for each request"},
			&cli.IntFlag{Name: "timeout", Value: 30, Usage: "stop match if timeout in second"},
		},
		Action: HttpMatch,
	}
	HTTP := cli.Command{
		Name:  "http",
		Usage: "http toolkit",
		Subcommands: []*cli.Command{
			&HttpMatch,
		},
	}
	app := &cli.App{
		Description: "DevOps Tools",
		Version:     "v0.0.1",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Value: false, Usage: "show more detail"},
		},
		Before: func(ctx *cli.Context) error {
			if ctx.Bool("verbose") {
				log.SetLevel(log.DebugLevel)
			}
			return nil
		},
		Commands: []*cli.Command{
			&HTTP,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
