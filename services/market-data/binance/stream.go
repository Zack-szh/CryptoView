package binance

// this package connects and reads data stream from binance websocket
import (
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

func StreamTickers(symbols []string, out chan<- TickerEvent) {
	url := buildURL(symbols, "@ticker")
	go func() {
		for {
			err := connect(url, func(data json.RawMessage) {
				var ticker TickerEvent
				if err := json.Unmarshal(data, &ticker); err != nil {
					log.Printf("failed to decode ticker: %s", err)
					return
				}
				out <- ticker
			})
			if err != nil {
				log.Printf("websocket error: %v — retrying in 5s", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func StreamBookTickers(symbols []string, out chan<- BookTickerEvent) {
	url := buildURL(symbols, "@bookTicker")
	go func() {
		for {
			err := connect(url, func(data json.RawMessage) {
				var bookticker BookTickerEvent
				if err := json.Unmarshal(data, &bookticker); err != nil {
					log.Printf("failed to decode book ticker: %s", err)
					return
				}
				out <- bookticker
			})
			if err != nil {
				log.Printf("websocket error: %v — retrying in 5s", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

// streamType is either @bookTicker or @ticker
func buildURL(symbols []string, streamType string) string {
	endpoints := make([]string, len(symbols))
	for i, s := range symbols {
		endpoints[i] = strings.ToLower(s) + streamType
	}
	return wsBaseURL + "/stream?streams=" + strings.Join(endpoints, "/")
}

// connects to websocket and calls handler for each message received
// handler just takes care of the specific json format for different endpoints
func connect(url string, handler func(json.RawMessage)) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	// cleanup when connection exits
	defer conn.Close()

	log.Printf("connected to websocket: %s", url)

	for {
		// we omit message_type here, always is JSON
		_, message, err := conn.ReadMessage()
		if err != nil {
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
