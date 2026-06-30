package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/szh/cryptoview/services/market-data/binance"
	"github.com/szh/cryptoview/services/market-data/db"
)

func main() {
	// first we connect to db, managed with context
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	ctx := context.Background()
	store, err := db.New(ctx, dbURL)

	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer store.Close()

	symbols := []string{"BTCUSDT"}

	tickerCh := make(chan binance.TickerEvent)
	bookTickerCh := make(chan binance.BookTickerEvent)
	tradeCh := make(chan binance.TradeEvent)

	binance.StreamTicker(symbols, tickerCh)
	binance.StreamBookTicker(symbols, bookTickerCh)
	binance.StreamTrade(symbols, tradeCh)

	fmt.Println("streaming — press Ctrl+C to stop")

	for {
		select {
		case event := <-tickerCh:
			fmt.Printf("[ticker]     [%s]  last=%-14s  change=%s%%\n",
				event.Symbol, event.LastPrice, event.PriceChangePct)
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
		}
	}
}
