package binance

// this package connects and reads data stream from binance websocket
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// for USA
const wsBaseURL = "wss://stream.binance.us:9443"

// for rest of the world
// const wsBaseURL = "wss://stream.binance.com:9443"

// Combined stream events are wrapped as follows: {"stream":"<streamName>","data":<rawPayload>}
type CombinedStream struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// 24hr rolling ticker statistics, pushed every ~1 second
type TickerEvent struct {
	EventType        string `json:"e"`
	EventTime        int64  `json:"E"`
	Symbol           string `json:"s"`
	PriceChange      string `json:"p"`
	PriceChangePct   string `json:"P"`
	WeightedAvgPrice string `json:"w"`
	FirstTradePrice  string `json:"x"`
	LastPrice        string `json:"c"`
	LastQty          string `json:"Q"`
	BestBidPrice     string `json:"b"`
	BestBidQty       string `json:"B"`
	BestAskPrice     string `json:"a"`
	BestAskQty       string `json:"A"`
	OpenPrice        string `json:"o"`
	High             string `json:"h"`
	Low              string `json:"l"`
	Volume           string `json:"v"`
	QuoteVolume      string `json:"q"`
	StatsOpenTime    int64  `json:"O"`
	StatsCloseTime   int64  `json:"C"`
	FirstTradeID     int64  `json:"F"`
	LastTradeID      int64  `json:"L"`
	TradeCount       int64  `json:"n"`
}

// best bid/ask from the order book, pushed instantly on every order book change
type BookTickerEvent struct {
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

type TradeEvent struct {
	EventType     string `json:"e"`
	EventTime     int64  `json:"E"`
	Symbol        string `json:"s"`
	TradeID       int64  `json:"t"`
	Price         string `json:"p"`
	Quantity      string `json:"q"`
	TradeTime     int64  `json:"T"`
	IsMarketMaker bool   `json:"m"`
	Ignore        bool   `json:"M"`
}

type KlineData struct {
	OpenTime     int64  `json:"t"`
	CloseTime    int64  `json:"T"`
	Symbol       string `json:"s"`
	Interval     string `json:"i"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"L"`
	OpenPrice    string `json:"o"`
	ClosePrice   string `json:"c"`
	High         string `json:"h"`
	Low          string `json:"l"`
	Volume       string `json:"v"`
	TradeCount   int64  `json:"n"`
	IsClosed     bool   `json:"x"`
}

type KlineEvent struct {
	EventType string    `json:"e"`
	EventTime int64     `json:"E"`
	Symbol    string    `json:"s"`
	Kline     KlineData `json:"k"`
}

// instead of three stream functions and repetitive code
// we can cover all streams using type parameters
func Stream[T any](ctx context.Context, symbols []string, streamType string, out chan<- T) {
	url := buildURL(symbols, streamType)
	go func() {
		for {
			err := connect(ctx, url, func(data json.RawMessage) {
				var event T
				if err := json.Unmarshal(data, &event); err != nil {
					// if fails to decode json object
					log.Printf("failed to decode json: %s", err)
					return
				}
				out <- event
			})

			// shutdown on context
			if ctx.Err() != nil {
				// context canceled, stop retrying, return
				return
			}
			// would only retry if we have not shutdown context
			if err != nil {
				log.Printf("websocket error: %v - retrying in 5s", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func StreamTicker(ctx context.Context, symbols []string, out chan<- TickerEvent) {
	Stream(ctx, symbols, "@ticker", out)
}

func StreamBookTicker(ctx context.Context, symbols []string, out chan<- BookTickerEvent) {
	Stream(ctx, symbols, "@bookTicker", out)
}

func StreamTrade(ctx context.Context, symbols []string, out chan<- TradeEvent) {
	Stream(ctx, symbols, "@trade", out)
}

func StreamKline(ctx context.Context, symbols []string, interval string, out chan<- KlineEvent) {
	Stream(ctx, symbols, "@kline_"+interval, out)
}

// streamType is either @ticker, @bookTicker, @trade
func buildURL(symbols []string, streamType string) string {
	endpoints := make([]string, len(symbols))
	for i, s := range symbols {
		endpoints[i] = strings.ToLower(s) + streamType
	}
	return wsBaseURL + "/stream?streams=" + strings.Join(endpoints, "/")
}

// connects to websocket and calls handler for each message received
// handler just takes care of the specific json format for different endpoints
// connect accepts context, closes the connectioon when ctx fires
func connect(ctx context.Context, url string, handler func(json.RawMessage)) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	// auto cleanup when connect() returns
	defer conn.Close()

	// active shutdown on context.Done()
	// this goroutine keeps listening on ctx for shutdown signals
	// IMPORTANT: CAUTIOUS OF LEAKING GOROUTINE!!!
	// this goroutine would be alive forever if a network failure occurs and connect returns
	// therefore we need a way to shutdown this goroutine when its parent process connect() exits
	done := make(chan struct{}) // empty signal channel
	defer close(done)
	go func() {
		select {
		// active shutdown on context
		case <-ctx.Done():
			conn.Close()
		// shutdown due to error
		// when connect() returns, defer close(done) is run
		// when done channel is closed, this goroutine select case <- done: immediately
		// therefore shutting down this goroutine
		case <-done:
		}
	}()

	log.Printf("connected to websocket: %s", url)

	for {
		// we omit message_type here, always is JSON
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				// if ctx.Err() is not nil, that means we actively shut down the context
				// therefore this is not a real error, simply return nil
				// clean exit
				return nil
			}
			// this is a real error
			// happens when we have not shut down context and it still fails to read message
			return fmt.Errorf("failed to read message: %w", err)
		}

		var envelope CombinedStream
		if err = json.Unmarshal(message, &envelope); err != nil {
			log.Printf("failed to decode envelope: %s", err)
			continue
		}

		handler(envelope.Data)
	}
}
