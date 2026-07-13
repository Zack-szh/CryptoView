import type {Ticker, Trade, Kline} from './types'

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

export async function fetchKline(symbol: string, interval = "1m", limit = 10): Promise<Kline[]> {
    const res = await fetch(`${BASE}/kline/${symbol}?interval=${interval}&limit=${limit}`)
    if (!res.ok) throw new Error("Failed to fetch kline")
    return res.json()
}