import type {Ticker, Trade, Kline, OrderBook} from './types'

const BASE = '/api/v1'

export async function fetchSymbol(): Promise<string[]> {
    const res = await fetch(`${BASE}/symbols`)
    if (!res.ok) throw new Error("Failed to fetch symbol")
    return res.json()
}

export async function fetchTicker(symbol: string, limit = 10): Promise<Ticker[]> {
    const res = await fetch(`${BASE}/ticker/${symbol}?limit=${limit}`)
    if (!res.ok) throw new Error("Failed to fetch ticker")
    return res.json()
}

export async function fetchTrade(symbol: string, limit = 10): Promise<Trade[]> {
    const res = await fetch(`${BASE}/trade/${symbol}?limit=${limit}`)
    if (!res.ok) throw new Error("Failed to fetch trade")
    return res.json()
}

export async function fetchKline(symbol: string, interval = "1m", sinceMs?: number): Promise<Kline[]> {
    const since = sinceMs ?? Date.now() - 30 * 24 * 60 * 60 * 1000
    const res = await fetch(`${BASE}/kline/${symbol}?interval=${interval}&since=${since}`)
    if (!res.ok) throw new Error("Failed to fetch kline")
    return res.json()
}

export async function fetchOrderBook(symbol: string, limit = 20): Promise<OrderBook> {
    const res = await fetch(`${BASE}/orderbook/${symbol}?limit=${limit}`)
    if (!res.ok) throw new Error("Failed to fetch orderbook")
    return res.json()
}