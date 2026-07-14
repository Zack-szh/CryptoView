import { useEffect, useRef } from 'react'
import { CandlestickSeries, ColorType, createChart } from 'lightweight-charts'
import type {
  CandlestickData,
  IChartApi,
  ISeriesApi,
  TickMarkFormatter,
  Time,
  UTCTimestamp,
} from 'lightweight-charts'
import type { Kline } from '../api/types'

interface Props {
  klines: Kline[]
  interval: string
  onIntervalChange: (i: string) => void
}

const INTERVALS = ['1m', '5m', '15m', '1h', '4h']

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

function toUtcTimestamp(openTime: string): UTCTimestamp | null {
  const milliseconds = Date.parse(openTime)

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

export default function KlineChart({ klines, interval, onIntervalChange }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const intervalRef = useRef(interval)
  const displayedIntervalRef = useRef<string | null>(null)

  useEffect(() => {
    if (!containerRef.current) return

    const tickMarkFormatter: TickMarkFormatter = (time) => formatAxisTime(time, intervalRef.current)
    const chart = createChart(containerRef.current, {
      layout: { background: { type: ColorType.Solid, color: '#111827' }, textColor: '#9ca3af' },
      grid: { vertLines: { color: '#1f2937' }, horzLines: { color: '#1f2937' } },
      width: containerRef.current.clientWidth,
      height: 320,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        tickMarkFormatter,
      },
      localization: {
        timeFormatter: formatHoverTime,
      },
    })

    const series = chart.addSeries(CandlestickSeries, {
      upColor: '#22c55e',
      downColor: '#ef4444',
      borderVisible: false,
      wickUpColor: '#22c55e',
      wickDownColor: '#ef4444',
    })
    chartRef.current = chart
    seriesRef.current = series

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
    const series = seriesRef.current
    if (!chart || !series) return

    const intervalKlines = klines.filter((kline) => kline.interval === interval)
    const data = normalizeKlines(intervalKlines)
    if (data.length === 0) return

    series.setData(data)

    if (displayedIntervalRef.current !== interval) {
      intervalRef.current = interval
      chart.timeScale().fitContent()
      displayedIntervalRef.current = interval
    }
  }, [interval, klines])

  function handleIntervalChange(nextInterval: string) {
    if (nextInterval === interval) return

    seriesRef.current?.setData([])
    displayedIntervalRef.current = null
    onIntervalChange(nextInterval)
  }

  return (
    <div>
      <div className="flex gap-2 mb-3">
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
      <div ref={containerRef} />
    </div>
  )
}
