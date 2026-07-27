import { useEffect, useRef, useState } from 'react'
import { CandlestickSeries, ColorType, createChart, HistogramSeries, LineSeries, LineStyle } from 'lightweight-charts'
import type {
  CandlestickData,
  IChartApi,
  ISeriesApi,
  LineData,
  TickMarkFormatter,
  Time,
  UTCTimestamp,
} from 'lightweight-charts'
import type { Kline, IndicatorSeries } from '../api/types'

interface Props {
  klines: Kline[]
  interval: string
  onIntervalChange: (i: string) => void
  series: IndicatorSeries | null
}

const INTERVALS = ['1m', '5m', '15m', '1h', '4h']

// dark-mode categorical slots, validated against this chart's #111827 surface
const COLOR_SMA20 = '#3987e5'
const COLOR_SMA50 = '#d95926'
const COLOR_EMA12 = '#199e70'
const COLOR_EMA26 = '#c98500'
const COLOR_VWAP = '#d55181'
const COLOR_BOLLINGER = '#9085e9'
const COLOR_MACD = '#3987e5'
const COLOR_MACD_SIGNAL = '#d95926'
const COLOR_RSI = '#c98500'
const COLOR_UP = '#22c55e'
const COLOR_DOWN = '#ef4444'

// each key toggles one indicator's series (Bollinger = mid+upper+lower together,
// MACD = macd+signal+histogram together, since those render as a group)
type OverlayKey = 'sma20' | 'sma50' | 'ema12' | 'ema26' | 'vwap' | 'bollinger'
type PaneKey = 'macd' | 'rsi'
type ToggleKey = OverlayKey | PaneKey

const OVERLAY_TOGGLES: { key: OverlayKey; label: string; color: string }[] = [
  { key: 'sma20', label: 'SMA 20', color: COLOR_SMA20 },
  { key: 'sma50', label: 'SMA 50', color: COLOR_SMA50 },
  { key: 'ema12', label: 'EMA 12', color: COLOR_EMA12 },
  { key: 'ema26', label: 'EMA 26', color: COLOR_EMA26 },
  { key: 'vwap', label: 'VWAP', color: COLOR_VWAP },
  { key: 'bollinger', label: 'Bollinger', color: COLOR_BOLLINGER },
]

const PANE_TOGGLES: { key: PaneKey; label: string; color: string }[] = [
  { key: 'macd', label: 'MACD', color: COLOR_MACD },
  { key: 'rsi', label: 'RSI 14', color: COLOR_RSI },
]

const ALL_TOGGLES = [...OVERLAY_TOGGLES, ...PANE_TOGGLES]

const axisTimeFormatter = new Intl.DateTimeFormat('en-US', {
  timeZone: 'UTC',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
})

const axisDateTimeFormatter = new Intl.DateTimeFormat('en-US', {
  timeZone: 'UTC',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  hour12: false,
})

const hoverTimeFormatter = new Intl.DateTimeFormat('en-US', {
  timeZone: 'UTC',
  year: 'numeric',
  month: 'short',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
  timeZoneName: 'short',
})

function toUtcTimestamp(isoTime: string): UTCTimestamp | null {
  const milliseconds = Date.parse(isoTime)

  return Number.isNaN(milliseconds)
    ? null
    : (Math.floor(milliseconds / 1000) as UTCTimestamp)
}

function toDate(time: Time): Date {
  if (typeof time === 'number') return new Date(time * 1000)

  if (typeof time === 'string') return new Date(time)

  return new Date(Date.UTC(time.year, time.month - 1, time.day))
}

function formatAxisTime(time: Time, interval: string): string {
  const date = toDate(time)

  return interval === '1h' || interval === '4h'
    ? axisDateTimeFormatter.format(date)
    : axisTimeFormatter.format(date)
}

function formatHoverTime(time: Time): string {
  return hoverTimeFormatter.format(toDate(time))
}

function normalizeKlines(klines: Kline[]): CandlestickData<UTCTimestamp>[] {
  const candlesByTime = new Map<UTCTimestamp, CandlestickData<UTCTimestamp>>()

  for (const kline of klines) {
    const time = toUtcTimestamp(kline.open_time)
    if (time === null) continue

    candlesByTime.set(time, {
      time,
      open: kline.open,
      high: kline.high,
      low: kline.low,
      close: kline.close,
    })
  }

  return [...candlesByTime.values()].sort((left, right) => left.time - right.time)
}

// zips IndicatorSeries.times against one indicator column, dropping points that
// aren't computable yet (null) — times/values are always index-aligned per the backend
function normalizeIndicatorLine(times: string[], values: (number | null)[]): LineData<UTCTimestamp>[] {
  const points: LineData<UTCTimestamp>[] = []

  for (let i = 0; i < times.length; i++) {
    const value = values[i]
    if (value === null || value === undefined) continue

    const time = toUtcTimestamp(times[i])
    if (time === null) continue

    points.push({ time, value })
  }

  return points.sort((left, right) => left.time - right.time)
}

type LinePaneSeries = ISeriesApi<'Line'>
type HistPaneSeries = ISeriesApi<'Histogram'>

export default function KlineChart({ klines, interval, onIntervalChange, series }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const intervalRef = useRef(interval)
  const displayedIntervalRef = useRef<string | null>(null)

  // main-pane overlays
  const sma20Ref = useRef<LinePaneSeries | null>(null)
  const sma50Ref = useRef<LinePaneSeries | null>(null)
  const ema12Ref = useRef<LinePaneSeries | null>(null)
  const ema26Ref = useRef<LinePaneSeries | null>(null)
  const vwapRef = useRef<LinePaneSeries | null>(null)
  const bollUpperRef = useRef<LinePaneSeries | null>(null)
  const bollMidRef = useRef<LinePaneSeries | null>(null)
  const bollLowerRef = useRef<LinePaneSeries | null>(null)

  // MACD sub-pane
  const macdRef = useRef<LinePaneSeries | null>(null)
  const macdSignalRef = useRef<LinePaneSeries | null>(null)
  const macdHistRef = useRef<HistPaneSeries | null>(null)

  // RSI sub-pane
  const rsiRef = useRef<LinePaneSeries | null>(null)

  const [visible, setVisible] = useState<Record<ToggleKey, boolean>>(() =>
    Object.fromEntries(ALL_TOGGLES.map((t) => [t.key, true])) as Record<ToggleKey, boolean>,
  )

  function toggleIndicator(key: ToggleKey) {
    setVisible((prev) => ({ ...prev, [key]: !prev[key] }))
  }

  useEffect(() => {
    if (!containerRef.current) return

    const tickMarkFormatter: TickMarkFormatter = (time) => formatAxisTime(time, intervalRef.current)
    const chart = createChart(containerRef.current, {
      layout: { background: { type: ColorType.Solid, color: '#111827' }, textColor: '#9ca3af' },
      grid: { vertLines: { color: '#1f2937' }, horzLines: { color: '#1f2937' } },
      width: containerRef.current.clientWidth,
      height: 520,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        tickMarkFormatter,
      },
      localization: {
        timeFormatter: formatHoverTime,
      },
    })

    const candles = chart.addSeries(CandlestickSeries, {
      upColor: COLOR_UP,
      downColor: COLOR_DOWN,
      borderVisible: false,
      wickUpColor: COLOR_UP,
      wickDownColor: COLOR_DOWN,
    })
    chartRef.current = chart
    seriesRef.current = candles

    // main pane (0) overlays — thin lines, no price line/last-value labels so they
    // don't compete visually with the candles
    const lineDefaults = { lastValueVisible: false, priceLineVisible: false, lineWidth: 1 as const }
    sma20Ref.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_SMA20 }, 0)
    sma50Ref.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_SMA50 }, 0)
    ema12Ref.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_EMA12 }, 0)
    ema26Ref.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_EMA26 }, 0)
    vwapRef.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_VWAP }, 0)
    bollMidRef.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_BOLLINGER }, 0)
    bollUpperRef.current = chart.addSeries(
      LineSeries,
      { ...lineDefaults, color: COLOR_BOLLINGER, lineStyle: LineStyle.Dashed },
      0,
    )
    bollLowerRef.current = chart.addSeries(
      LineSeries,
      { ...lineDefaults, color: COLOR_BOLLINGER, lineStyle: LineStyle.Dashed },
      0,
    )

    // pane 1: MACD
    macdHistRef.current = chart.addSeries(HistogramSeries, { lastValueVisible: false, priceLineVisible: false }, 1)
    macdRef.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_MACD }, 1)
    macdSignalRef.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_MACD_SIGNAL }, 1)

    // pane 2: RSI, with 30/70 reference lines
    rsiRef.current = chart.addSeries(LineSeries, { ...lineDefaults, color: COLOR_RSI }, 2)
    rsiRef.current.createPriceLine({
      price: 70,
      color: '#4b5563',
      lineWidth: 1,
      lineStyle: LineStyle.Dotted,
      axisLabelVisible: true,
      title: '',
    })
    rsiRef.current.createPriceLine({
      price: 30,
      color: '#4b5563',
      lineWidth: 1,
      lineStyle: LineStyle.Dotted,
      axisLabelVisible: true,
      title: '',
    })

    const panes = chart.panes()
    if (panes[1]) panes[1].setHeight(120)
    if (panes[2]) panes[2].setHeight(100)

    const resizeObserver = new ResizeObserver(() => {
      chart.applyOptions({ width: containerRef.current!.clientWidth })
    })
    resizeObserver.observe(containerRef.current)

    return () => {
      resizeObserver.disconnect()
      chart.remove()
      chartRef.current = null
      seriesRef.current = null
    }
  }, [])

  useEffect(() => {
    const chart = chartRef.current
    const candles = seriesRef.current
    if (!chart || !candles) return

    const intervalKlines = klines.filter((kline) => kline.interval === interval)
    const data = normalizeKlines(intervalKlines)
    if (data.length === 0) return

    candles.setData(data)

    if (displayedIntervalRef.current !== interval) {
      intervalRef.current = interval
      chart.timeScale().fitContent()
      displayedIntervalRef.current = interval
    }
  }, [interval, klines])

  useEffect(() => {
    if (!series || series.interval !== interval) return

    sma20Ref.current?.setData(normalizeIndicatorLine(series.times, series.sma_20))
    sma50Ref.current?.setData(normalizeIndicatorLine(series.times, series.sma_50))
    ema12Ref.current?.setData(normalizeIndicatorLine(series.times, series.ema_12))
    ema26Ref.current?.setData(normalizeIndicatorLine(series.times, series.ema_26))
    vwapRef.current?.setData(normalizeIndicatorLine(series.times, series.vwap))
    bollMidRef.current?.setData(normalizeIndicatorLine(series.times, series.bollinger_mid))
    bollUpperRef.current?.setData(normalizeIndicatorLine(series.times, series.bollinger_upper))
    bollLowerRef.current?.setData(normalizeIndicatorLine(series.times, series.bollinger_lower))

    macdRef.current?.setData(normalizeIndicatorLine(series.times, series.macd))
    macdSignalRef.current?.setData(normalizeIndicatorLine(series.times, series.macd_signal))
    macdHistRef.current?.setData(
      normalizeIndicatorLine(series.times, series.macd_hist).map((point) => ({
        ...point,
        color: point.value >= 0 ? COLOR_UP : COLOR_DOWN,
      })),
    )

    rsiRef.current?.setData(normalizeIndicatorLine(series.times, series.rsi_14))
  }, [series, interval])

  useEffect(() => {
    sma20Ref.current?.applyOptions({ visible: visible.sma20 })
    sma50Ref.current?.applyOptions({ visible: visible.sma50 })
    ema12Ref.current?.applyOptions({ visible: visible.ema12 })
    ema26Ref.current?.applyOptions({ visible: visible.ema26 })
    vwapRef.current?.applyOptions({ visible: visible.vwap })
    bollMidRef.current?.applyOptions({ visible: visible.bollinger })
    bollUpperRef.current?.applyOptions({ visible: visible.bollinger })
    bollLowerRef.current?.applyOptions({ visible: visible.bollinger })

    macdRef.current?.applyOptions({ visible: visible.macd })
    macdSignalRef.current?.applyOptions({ visible: visible.macd })
    macdHistRef.current?.applyOptions({ visible: visible.macd })

    rsiRef.current?.applyOptions({ visible: visible.rsi })
  }, [visible])

  function handleIntervalChange(nextInterval: string) {
    if (nextInterval === interval) return

    seriesRef.current?.setData([])
    displayedIntervalRef.current = null
    onIntervalChange(nextInterval)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <div className="flex gap-2">
          {INTERVALS.map((i) => (
            <button
              key={i}
              onClick={() => handleIntervalChange(i)}
              className={`px-3 py-1 text-xs rounded transition-colors ${
                i === interval ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {i}
            </button>
          ))}
        </div>
        <div className="flex flex-wrap gap-x-2 gap-y-1">
          {OVERLAY_TOGGLES.map(({ key, label, color }) => (
            <button
              key={key}
              onClick={() => toggleIndicator(key)}
              className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs transition-opacity hover:bg-gray-800 ${
                visible[key] ? 'text-gray-300 opacity-100' : 'text-gray-500 opacity-40'
              }`}
            >
              <span className="inline-block w-2.5 h-0.5" style={{ backgroundColor: color }} />
              {label}
            </button>
          ))}
        </div>
      </div>
      <div ref={containerRef} />
      <div className="flex flex-wrap gap-x-2 gap-y-1 mt-1">
        {PANE_TOGGLES.map(({ key, label, color }) => (
          <button
            key={key}
            onClick={() => toggleIndicator(key)}
            className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs transition-opacity hover:bg-gray-800 ${
              visible[key] ? 'text-gray-400 opacity-100' : 'text-gray-500 opacity-40'
            }`}
          >
            <span className="inline-block w-2.5 h-0.5" style={{ backgroundColor: color }} />
            {label}
          </button>
        ))}
      </div>
    </div>
  )
}
