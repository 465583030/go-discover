package main

import (
	"log"
	"time"

	discover "github.com/skilstak/go-discover"
)

func main() {
	start := time.Now()
	crawler := discover.NewCrawler()

	crawler.Name = "Example"
	crawler.IPChunkSize = 10
	crawler.FilesToLookFor = []string{""}
	crawler.GoRoutines = 10

	crawler.Discover()

	elapsed := time.Since(start)

	log.Printf("Program took %s", elapsed)
}
