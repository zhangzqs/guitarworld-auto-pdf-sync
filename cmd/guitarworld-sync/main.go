package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/zhangzqs/guitarworld-auto-pdf-sync/internal/models"
	"github.com/zhangzqs/guitarworld-auto-pdf-sync/internal/scraper"
)

var (
	outputDir      string
	maxConcurrency int
)

func main() {
	flag.StringVar(&outputDir, "output", "./pdfs", "Output directory for PDFs")
	flag.IntVar(&maxConcurrency, "concurrency", 3, "Maximum number of concurrent downloads")
	flag.Parse()

	cookies := os.Getenv("GUITARWORLD_COOKIES")
	if cookies == "" {
		log.Fatal("Please set GUITARWORLD_COOKIES environment variable")
	}

	xsrfToken := scraper.ExtractXSRFToken(cookies)
	if xsrfToken == "" {
		log.Println("Warning: XSRF-TOKEN not found in cookies, requests may fail")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	sc := scraper.New(cookies, xsrfToken)
	processor := scraper.NewProcessor(sc, outputDir)

	log.Println("Fetching sheet music list from API...")
	sheets, err := sc.FetchAllSheetMusic()
	if err != nil {
		log.Fatalf("Failed to fetch sheet music list: %v", err)
	}

	log.Printf("\n=== Found %d sheet music items ===\n", len(sheets))
	log.Printf("Using concurrency: %d\n", maxConcurrency)

	// 使用并发处理
	results := make(chan models.ProcessResult, len(sheets))
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, sheet := range sheets {
		wg.Add(1)
		go func(s models.SheetMusicItem, idx int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := processor.ProcessSheet(s, idx+1, len(sheets))
			results <- result

			time.Sleep(1 * time.Second)
		}(sheet, i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	successCount := 0
	errorCount := 0
	skippedCount := 0

	for result := range results {
		if result.Success {
			if result.Skipped {
				skippedCount++
			} else {
				successCount++
			}
		} else {
			errorCount++
			if result.Error != nil {
				log.Printf("Error processing %s: %v", result.Sheet.Name, result.Error)
			}
		}
	}

	log.Printf("\n=== Summary ===")
	log.Printf("Total: %d", len(sheets))
	log.Printf("Success: %d", successCount)
	log.Printf("Skipped: %d", skippedCount)
	log.Printf("Errors: %d", errorCount)
}
