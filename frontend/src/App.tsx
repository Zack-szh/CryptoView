import { useState, useEffect, useCallback } from 'react'
import { fetchSymbol, fetchTicker, fetchTrade, fetchKline, fetchOrderBook, fetchIndicator } from './api/client'
import type { Ticker, Trade, Kline, OrderBook, Indicator } from './api/types'
import SymbolSelector from './components/SymbolSelector'
import TickerPanel from './components/TickerPanel'
import TradeTable from './components/TradeTable'
import KlineChart from './components/KlineChart'
import OrderBookTable from './components/OrderBookTable'
import OrderBookDepth from './components/OrderBookDepth'
import IndicatorPanel from './components/IndicatorPanel'

export default function App() {
  const [symbols, setSymbols] = useState<string[]>([])
  const [selected, setSelected] = useState<string>('')
  const [ticker, setTicker] = useState<Ticker | null>(null)
  const [trades, setTrades] = useState<Trade[]>([])
  const [klines, setKlines] = useState<Kline[]>([])
  // NOTE: setInterval is reserved for javascript, thus setInter
  const [inter, setInter] = useState('1m')
  const [error, setError] = useState<string | null>(null)
  const [orderBook, setOrderBook] = useState<OrderBook | null>(null)
  const [indicator, setIndicator] = useState<Indicator | null>(null)

  useEffect(() => {
    fetchSymbol()
      .then((s) => { setSymbols(s); setSelected(s[0] ?? '') })
      .catch(() => setError('Could not reach API'))
  }, [])

  const refresh = useCallback(() => {
    if (!selected) return
    Promise.all([
      fetchTicker(selected, 1).then((d) => setTicker(d[0] ?? null)),
      fetchTrade(selected, 30).then(setTrades),
    ]).catch(() => setError('Fetch error'))
  }, [selected])

  const refreshKlines = useCallback(() => {
    if (!selected) return
    fetchKline(selected, inter).then(setKlines).catch(() => {})
  }, [selected, inter])
  // refreshKlines is a separate refresh because we can select different time frame (1m, 5m, etc)
  // and everytime we choose a different interval it should refresh

  // for now we fetch order book depth = 20
  const refreshOrderBook = useCallback(() => {
    if (!selected) return
    fetchOrderBook(selected, 20).then(setOrderBook).catch(() => {})
  }, [selected])

  const refreshIndicator = useCallback(() => {
    if (!selected) return 
    fetchIndicator(selected, inter).then(setIndicator).catch(() => {})
  }, [selected, inter])

  useEffect(() => {
    refresh()
    refreshKlines()
    refreshOrderBook()
    refreshIndicator()
    
    const t1 = setInterval(refresh, 2000)
    const t2 = setInterval(refreshKlines, 2000)
    const t3 = setInterval(refreshOrderBook, 1000)
    const t4 = setInterval(refreshIndicator, 2000)
    return () => { clearInterval(t1); clearInterval(t2); clearInterval(t3); clearInterval(t4) }
  }, [refresh, refreshKlines, refreshOrderBook, refreshIndicator])

  return (
    <div className="min-h-screen bg-gray-950 text-white p-6 space-y-6">
      <header className="flex items-center justify-between">
        <h1 className="text-xl font-bold tracking-tight">CryptoView</h1>
        {error && <span className="text-red-400 text-sm">{error}</span>}
      </header>

      <SymbolSelector symbols={symbols} selected={selected} onChange={setSelected} />

      <section className="bg-gray-900 rounded-xl p-5">
        <TickerPanel ticker={ticker} />
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <section className="bg-gray-900 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Recent Trades (binance.us)</h2>
          <TradeTable trades={trades} />
        </section>

        <section className="bg-gray-900 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Candlestick Chart</h2>
          <KlineChart klines={klines} interval={inter} onIntervalChange={setInter} />
        </section>

        <section className="bg-gray-900 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Indicators</h2>
          <IndicatorPanel indicators={indicator} /> 
        </section>

         <section className="bg-gray-900 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Order Book (RAW)</h2>
          <OrderBookTable book={orderBook} />
        </section>

        <section className="bg-gray-900 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-gray-400 mb-3 uppercase tracking-wider">Order Book (DEPTH)</h2>
          <OrderBookDepth book={orderBook} />
        </section>

      </div>
    </div>
  )
}