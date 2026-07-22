import type { Indicator } from '../api/types'

interface Props {
    indicators: Indicator | null 
}

// formating for displaying float64
function fmt(n: number | null | undefined, digits = 2) {
    if (n == null || n == undefined) return '-'

    return n.toLocaleString('en-US', {minimumFractionDigits: digits, maximumFractionDigits: digits})
}

export default function IndicatorPanel({ indicators }: Props) {
    if (!indicators) return <div className="text-gray-500 text-sm">Loading stats…</div>

    const rsi = indicators.rsi_14
    const rsiColor =
    rsi === null || rsi === undefined ? 'text-white'
    : rsi >= 70 ? 'text-red-400'
    : rsi <= 30 ? 'text-green-400'
    : 'text-white'

    const histColor = (indicators.macd?.histogram ?? 0) >= 0 ? 'text-green-400' : 'text-red-400'
    const imbColor = (indicators.orderbook_imbalance ?? 0) >= 0 ? 'text-green-400' : 'text-red-400'

    return (
    <div className="grid grid-cols-2 gap-x-8 gap-y-1 text-sm text-gray-400">
      <span>SMA 20 <span className="text-white">${fmt(indicators.sma_20)}</span></span>
      <span>SMA 50 <span className="text-white">${fmt(indicators.sma_50)}</span></span>
      <span>EMA 12 <span className="text-white">${fmt(indicators.ema_12)}</span></span>
      <span>EMA 26 <span className="text-white">${fmt(indicators.ema_26)}</span></span>
      <span>RSI 14 <span className={rsiColor}>{fmt(indicators.rsi_14)}</span></span>
      <span>MACD <span className="text-white">{fmt(indicators.macd?.macd, 4)}</span></span>
      <span>MACD Signal <span className="text-white">{fmt(indicators.macd?.signal, 4)}</span></span>
      <span>MACD Hist <span className={histColor}>{fmt(indicators.macd?.histogram, 4)}</span></span>
      <span>Bollinger Mid <span className="text-white">${fmt(indicators.bollinger?.middle)}</span></span>
      <span>Bollinger Upper <span className="text-white">${fmt(indicators.bollinger?.upper)}</span></span>
      <span>Bollinger Lower <span className="text-white">${fmt(indicators.bollinger?.lower)}</span></span>
      <span>Realized Vol (ann.) <span className="text-white">{fmt((indicators.realized_volatility ?? 0) * 100)}%</span></span>
      <span>VWAP <span className="text-white">${fmt(indicators.vwap)}</span></span>
      <span>OB Imbalance <span className={imbColor}>{fmt(indicators.orderbook_imbalance, 3)}</span></span>
    </div>
  )
}
