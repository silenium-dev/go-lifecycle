package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/silenium-dev/go-lifecycle/pkg/lifecycle"
	"log"
	"net/http"
	"strconv"
	"time"
)

var logCh = make(chan string)

func main() {
	app := lifecycle.NewApplication(&mainApp{})
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

type mainApp struct {
}

func (a *mainApp) ImmediateExit() {
	log.Println("immediate exit")
}

func (a *mainApp) Cleanup(ctx context.Context, loggingCtx context.Context) {
	if loggingCtx.Err() != nil {
		log.Println("logging context canceled")
	} else {
		logCh <- "cleanup"
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.randomnumberapi.com/api/v1.0/randomnumber", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var nums []int
	err = json.NewDecoder(resp.Body).Decode(&nums)
	if err != nil {
		log.Fatal(err)
	}
	logCh <- fmt.Sprintf("random number from randomnumberapi.com: %d", nums[0])
}

func (a *mainApp) Main(ctx context.Context, loggingCtx context.Context) error {
	go func() {
		for {
			select {
			case <-loggingCtx.Done():
				close(logCh)
				return
			case msg := <-logCh:
				println(msg)
			}
		}
	}()
	for j := 0; j < 10; j++ {
		select {
		case <-time.After(time.Second):
			logCh <- "tick"
		case <-ctx.Done():
			for i := 0; i < 4; i++ {
				logCh <- strconv.Itoa(i)
				select {
				case <-time.After(time.Second):
					break
				case <-loggingCtx.Done():
					return nil
				}
			}
			return nil
		}
	}
	return nil
}
