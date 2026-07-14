import { useState, useEffect, useCallback } from 'react'
import { fetchSymbol, fetchTicker, fetchTrade, fetchKline } from './api/client'
import type { Ticker, Trade, Kline } from './api/types'
import SymbolSelector from './components/SymbolSelector'
import TickerPanel from './components/TickerPanel'
import TradeTable from './components/TradeTable'
import KlineChart from './components/KlineChart'

export default function App() {
  const [symbols, setSymbols] = useState<string[]>([])
  const [selected, setSelected] = useState<string>('')
  const [ticker, setTicker] = useState<Ticker | null>(null)
  const [trades, setTrades] = useState<Trade[]>([])
  const [klines, setKlines] = useState<Kline[]>([])
  // NOTE: setInterval is reserved for javascript, thus setInter
  const [inter, setInter] = useState('1m')
  const [error, setError] = useState<string | null>(null)

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

  useEffect(() => {
    refresh()
    refreshKlines()
    // refresh data every 5 seconds
    const t1 = setInterval(refresh, 5000)
    const t2 = setInterval(refreshKlines, 5000)
    return () => { clearInterval(t1); clearInterval(t2) }
  }, [refresh, refreshKlines])

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
      </div>
    </div>
  )
}