package main

import (
	"fmt"

	"github.com/szh/cryptoview/market-data/binance"
)

func main() {
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
			fmt.Printf("[trade] [%s] price=%-14s quantity=%d\n",
				event.Symbol, event.Price, event.Quantity)
		}
	}
}
