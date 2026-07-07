package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/szh/cryptoview/services/market-data/binance"
	"github.com/szh/cryptoview/services/market-data/db"
)

func main() {
	// build context for signal-aware exit
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// first we connect to db, managed with context
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	store, err := db.New(ctx, dbURL)

	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer store.Close()

	symbols := strings.Split(os.Getenv("SYMBOLS"), ",")

	// stream events communication channels
	tickerCh := make(chan binance.TickerEvent)
	bookTickerCh := make(chan binance.BookTickerEvent)
	tradeCh := make(chan binance.TradeEvent)

	binance.StreamTicker(ctx, symbols, tickerCh)
	binance.StreamBookTicker(ctx, symbols, bookTickerCh)
	binance.StreamTrade(ctx, symbols, tradeCh)

	fmt.Println("streaming — press Ctrl+C to stop")

	// Change Log 7/7/2026
	// in the select statement below, insertTrade and insertTicker are synchronous, they might block
	// therefore we are going to put each insert in its own goroutine, with its own channel
	// to solve the go shutdown problem, we use the idiomatic go approach
	// when we close the goroutine, we close channels when goroutine terminate,
	// then wait for worker to drain the buffered channel of TickerEvent and TradeEvent
	// two options here:
	// 		1. Having a fixed number of goroutines, for example 4 workers competing to drain
	// 			tickerJobs and tradeJobs
	//		2. Having every single insert job launches its own goroutine.
	// 			this gives unbounded concurrency, but here we are bottlenecked by db, so we choose option1
	// we use a waitgroup to track and manage the number of goroutines launched

	tickerJobs := make(chan binance.TickerEvent, 500)
	tradeJobs := make(chan binance.TradeEvent, 500)

	var wg sync.WaitGroup

	// TODO!!!
	// currently every single job are being inserted one by one
	// later we might implement batch insert on pgx.pool for better performance

	for range 4 { // 4 workers for each channel, 8 workers launched in total
		wg.Add(2) // add 2, one trade worker one ticker worker
		go func() {
			defer wg.Done()
			for event := range tickerJobs {
				// insert tickerEvent into db
				// IMPORTANT!!!
				// do not pass the canceled context here, because we still need to
				// drain the channel even after we have called ctx.Done()
				// however we stil need to pass context because pgx requires it
				// so we just pass in "empty" context
				// fmt.Printf("DEBUG: inserting into tickers")
				if err := store.InsertTicker(context.Background(), event); err != nil {
					log.Printf("failed to insert ticker: %v", err)
				}
			}
		}()

		go func() {
			defer wg.Done()
			for event := range tradeJobs {
				// insert tradeEvent into db
				// fmt.Printf("DEBUG: inserting into trades")
				if err := store.InsertTrade(context.Background(), event); err != nil {
					log.Printf("failed to insert trade: %v", err)
				}
			}
		}()
	}

	for {
		select {
		case event := <-tickerCh:
			fmt.Printf("[ticker]     [%s]  last=%-14s  change=%s%%\n",
				event.Symbol, event.LastPrice, event.PriceChangePct)
			tickerJobs <- event
		case event := <-bookTickerCh:
			fmt.Printf("[bookTicker] [%s]  bid=%-14s  ask=%s\n",
				event.Symbol, event.BidPrice, event.AskPrice)
		case event := <-tradeCh:
			fmt.Printf("[trade] [%s] price=%-14s quantity=%s\n",
				event.Symbol, event.Price, event.Quantity)
			// insert trade into db
			tradeJobs <- event
		case <-ctx.Done():
			log.Printf("shutting down on context...")
			log.Printf("closing tickerJobs and tradeJobs channels...")
			close(tickerJobs)
			close(tradeJobs)
			// wait for workers to finish draining channels
			wg.Wait()
			return
		}
	}
}
