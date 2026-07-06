package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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

	tickerCh := make(chan binance.TickerEvent)
	bookTickerCh := make(chan binance.BookTickerEvent)
	tradeCh := make(chan binance.TradeEvent)

	binance.StreamTicker(ctx, symbols, tickerCh)
	binance.StreamBookTicker(ctx, symbols, bookTickerCh)
	binance.StreamTrade(ctx, symbols, tradeCh)

	fmt.Println("streaming — press Ctrl+C to stop")

	for {
		select {
		case event := <-tickerCh:
			fmt.Printf("[ticker]     [%s]  last=%-14s  change=%s%%\n",
				event.Symbol, event.LastPrice, event.PriceChangePct)
			// insert ticker into db
			if err := store.InsertTicker(ctx, event); err != nil {
				log.Printf("failed to insert ticker: %v", err)
			}
		case event := <-bookTickerCh:
			fmt.Printf("[bookTicker] [%s]  bid=%-14s  ask=%s\n",
				event.Symbol, event.BidPrice, event.AskPrice)
		case event := <-tradeCh:
			fmt.Printf("[trade] [%s] price=%-14s quantity=%s\n",
				event.Symbol, event.Price, event.Quantity)
			// insert trade into db
			if err := store.InsertTrade(ctx, event); err != nil {
				log.Printf("failed to insert trade: %v", err)
			}
		case <-ctx.Done():
			log.Printf("shutting down on context...")
			return
		}
	}
}
