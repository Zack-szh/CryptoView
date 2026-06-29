package main

import (
	"fmt"

	"github.com/szh/cryptoview/market-data/binance"
)

func main() {
	symbols := []string{"BTCUSDT"}

	tickers := make(chan binance.TickerEvent)
	bookTickers := make(chan binance.BookTickerEvent)

	binance.StreamTickers(symbols, tickers)
	binance.StreamBookTickers(symbols, bookTickers)

	fmt.Println("streaming — press Ctrl+C to stop")
	for {
		select {
		case e := <-tickers:
			fmt.Printf("[ticker]     [%s]  last=%-14s  change=%s%%\n",
				e.Symbol, e.LastPrice, e.PriceChangePct)
		case e := <-bookTickers:
			fmt.Printf("[bookTicker] [%s]  bid=%-14s  ask=%s\n",
				e.Symbol, e.BidPrice, e.AskPrice)
		}
	}
}
