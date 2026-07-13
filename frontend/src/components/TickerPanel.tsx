import type { Ticker } from '../api/types'

interface Props {
  ticker: Ticker | null
}

function fmt(n: number) {
  return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export default function TickerPanel({ ticker }: Props) {
  if (!ticker) return <div className="text-gray-500 text-sm">Loading ticker…</div>

  const change = ticker.last_price - ticker.open_price
  const changePct = (change / ticker.open_price) * 100
  const up = change >= 0

  return (
    <div className="flex flex-wrap gap-6 items-end">
      <div>
        <div className="text-3xl font-bold text-white">${fmt(ticker.last_price)}</div>
        <div className={`text-sm mt-1 ${up ? 'text-green-400' : 'text-red-400'}`}>
          {up ? '▲' : '▼'} {fmt(Math.abs(change))} ({changePct.toFixed(2)}%)
        </div>
      </div>
      <div className="grid grid-cols-2 gap-x-8 gap-y-1 text-sm text-gray-400">
        <span>24h High <span className="text-white">${fmt(ticker.high)}</span></span>
        <span>24h Low <span className="text-white">${fmt(ticker.low)}</span></span>
        <span>Volume <span className="text-white">{fmt(ticker.volume)}</span></span>
        <span>Trades <span className="text-white">{ticker.trade_count.toLocaleString()}</span></span>
      </div>
    </div>
  )
}