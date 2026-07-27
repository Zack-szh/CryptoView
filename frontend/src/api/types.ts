export interface Ticker {
  time: string
  symbol: string
  last_price: number
  open_price: number
  high: number
  low: number
  volume: number
  quote_volume: number
  weighted_avg_price: number
  trade_count: number
}

export interface Trade {
  time: string
  symbol: string
  price: number
  quantity: number
  is_maker: boolean
  trade_id: number
}

export interface Kline {
  open_time: string
  close_time: string
  symbol: string
  interval: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  trade_count: number
  is_closed: boolean
}

export interface OrderBook {
  last_update_id: number 
  bids: [string, string][]
  asks: [string, string][]
}

export interface MACDResult {
  macd: number
  signal: number
  histogram: number
}

export interface BollingerResult {
  middle: number
  upper: number
  lower: number
}

export interface Indicator {
  symbol: string
  interval: string
  time: string
  last_price: number
  sma_20: number | null
  sma_50: number | null
  ema_12: number | null
  ema_26: number | null
  rsi_14: number | null
  macd: MACDResult | null
  bollinger: BollingerResult | null
  realized_volatility: number | null
  vwap: number | null
  orderbook_imbalance: number | null
}

// each array below is index-aligned with `times`; entries are null until that indicator
// has enough history to compute (e.g. sma_50 is null for the first 49 points)
export interface IndicatorSeries {
  symbol: string
  interval: string
  times: string[]

  sma_20: (number | null)[]
  sma_50: (number | null)[]
  ema_12: (number | null)[]
  ema_26: (number | null)[]

  bollinger_mid: (number | null)[]
  bollinger_upper: (number | null)[]
  bollinger_lower: (number | null)[]

  vwap: (number | null)[]

  macd: (number | null)[]
  macd_signal: (number | null)[]
  macd_hist: (number | null)[]

  rsi_14: (number | null)[]
}