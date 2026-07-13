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