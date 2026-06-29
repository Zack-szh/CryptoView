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

// default binance stream endpoint
const wsBaseURL = "wss://stream.binance.com:9443"

// we are to use a combined-stream approach
// this is because I want to have separate stream for different ticker
// ex: BTDUSD has its stream, ETHUSD has its own stream
// Combined stream events are wrapped as follows: {"stream":"<streamName>","data":<rawPayload>}
type CombinedStream struct {
	Stream		string 	`json:"stream"`
	Data 		json.RawMessage 	`json:"data"`
}

// this is the data structure for each streamed message from binanace websocket, per spec
// documentation: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams
// we might not need all fields here but I will include all for now

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

// given a slice of ticker symbols, builds endpoints for all symbols and store in streamEndpoints[]
// a goroutine would connect to these endpoints, then stream data
// then pushes the data into a send-only channel of TickerEvent
// note that all symbols share the same channel, ONE CHANNEL FOR ALL TICKERS!
func StreamTickers(symbols []string, out chan<- TickerEvent) {
	streamEndpoints := make([]string, len(symbols))

	// build endpoint for each symbol
	for i, s := range(symbols) {
		streamEndpoints[i] = strings.ToLower(s) + "@ticker"
	}

	// then build the actual url to pull data from
	// url example: 
	// "wss://stream.binance.com:9443/stream?streams=btcusdt@ticker/ethusdt@ticker/solusdt@ticker"
	url := wsBaseURL + "/stream?streams=" + strings.Join(streamEndpoints, "/")	

	// acutal goroutine that runs connect() to connect to websocket
	// and return data into out channel, has error handling and retry logic
	go func(

	)
}

// conects to websocket and stream data
func connect(url string, out chan<- TickerEvent) error {

	// open client-side websocket connection 
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)

	// if connnection fails
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	// connection cleanup when connect() exits
	defer conn.close()

	log.Printf("connected to websocket: %s", url)

	// connection done, now streaming data

	for {
		// we omit message_type here, always is JSON
		_, message, err := conn.ReadMessage()

		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		// decode json
		var envelope CombinedStream
		err := json.Unmarshal(message, &envelope)	// pass by pointer, need to modify it

		if err != nil {
			log.Printf("failed to decode message: %s", err)
			continue
		}

		// decode ticker in envelope
		var ticker TickerEvent
		err := json.Unmarshal(envelope.Data, &ticker)

		if err != nil {
			log.Printf("failed to decode ticker: %s", err)
			continue 
		}

		// after decoding ticker, send to out channel
		out <- ticker
	}
}