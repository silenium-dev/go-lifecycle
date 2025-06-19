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
	app := lifecycle.NewApplication(appMain, appCleanup, nil)
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func appCleanup(ctx context.Context, loggingCtx context.Context) {
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

func appMain(ctx context.Context, loggingCtx context.Context) error {
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
	for {
		select {
		case <-time.After(time.Second):
			logCh <- "tick"
		case <-ctx.Done():
			for i := 0; i < 4; i++ {
				logCh <- strconv.Itoa(i)
				<-time.After(time.Second)
			}
			return nil
		}
	}
}
