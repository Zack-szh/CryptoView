import { useEffect, useRef } from 'react'
import { createChart, CandlestickSeries, ColorType } from 'lightweight-charts'
import type { IChartApi, ISeriesApi } from 'lightweight-charts'
import type { Kline } from '../api/types'

interface Props {
  klines: Kline[]
  interval: string
  onIntervalChange: (i: string) => void
}

const INTERVALS = ['1m', '5m', '15m', '1h', '4h']

export default function KlineChart({ klines, interval, onIntervalChange }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)

  useEffect(() => {
    if (!containerRef.current) return
    const chart = createChart(containerRef.current, {
      layout: { background: { type: ColorType.Solid, color: '#111827' }, textColor: '#9ca3af' },
      grid: { vertLines: { color: '#1f2937' }, horzLines: { color: '#1f2937' } },
      width: containerRef.current.clientWidth,
      height: 320,
      timeScale: {timeVisible: true, secondsVisible: false},
      localization: {
        timeFormatter: (ts: number) => 
          new Date(ts * 1000).toLocaleTimeString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            hour12: false,
          }),
      }
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

    const ro = new ResizeObserver(() => {
      chart.applyOptions({ width: containerRef.current!.clientWidth })
    })
    ro.observe(containerRef.current)

    return () => {
      ro.disconnect()
      chart.remove()
    }
  }, [])

  useEffect(() => {
    if (!seriesRef.current || klines.length === 0) return
    const data = klines.map((k) => ({
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      time: (Math.floor(new Date(k.open_time).getTime() / 1000)) as any,
      open: k.open,
      high: k.high,
      low: k.low,
      close: k.close,
    }))
    seriesRef.current.setData(data)
    chartRef.current?.timeScale().fitContent()
  }, [klines])

  return (
    <div>
      <div className="flex gap-2 mb-3">
        {INTERVALS.map((i) => (
          <button
            key={i}
            onClick={() => onIntervalChange(i)}
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