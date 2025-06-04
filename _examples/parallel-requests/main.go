// Copyright 2022 Fastly, Inc.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

func main() {
	fsthttp.ServeFunc(func(ctx context.Context, w fsthttp.ResponseWriter, r *fsthttp.Request) {
		// Log to the console (`fastly logs tail`) and the client.
		log := log.New(io.MultiWriter(os.Stdout, w), "", log.Ltime)
		log.Printf("Starting")
		begin := time.Now()

		// Send several requests in parallel.
		var wg sync.WaitGroup
		urls := []string{
			"https://http-me.glitch.me/drip=2?wait=3000", // 5s
			"https://http-me.glitch.me/drip=2?wait=2000", // 4s
			"https://http-me.glitch.me/wait=3000",        // 3s
		}

		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				fmt.Println("[goroutine] STARTING:", url)
				defer wg.Done()

				log.Printf("Starting %s", url)

				req, err := fsthttp.NewRequest(fsthttp.MethodGet, url, nil)
				if err != nil {
					log.Printf("%s: create request: %v", url, err)
					return
				}
				req.CacheOptions.Pass = true

				resp, err := req.Send(ctx, "httpme")
				if err != nil {
					log.Printf("%s: send request: %v", url, err)
					return
				}

				_, err = io.Copy(io.Discard, resp.Body)
				if err != nil {
					log.Printf("%s: stream response body: %v", url, err)
					return
				}

				fmt.Println("[goroutine] FINISHED:", url)
				log.Printf("Finished %s", url)
			}(url)
		}

		fmt.Println("[main] Waiting for all goroutines...")
		wg.Wait()
		fmt.Println("[main] All goroutines finished")
		log.Printf("Finished after %s", time.Since(begin))

		// All requests should finish in about as long as the longest individual
		// request took. That is, about 5s, rather than 5s+4s+3s=12s.
		log.Printf("Finished after %s", time.Since(begin))
	})
}
