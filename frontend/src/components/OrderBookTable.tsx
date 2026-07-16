import type { OrderBook } from '../api/types'

interface Props {
  book: OrderBook | null
}

function Row({
  price,
  quantity,
  side,
  pct,
}: {
  price: string
  quantity: string
  side: 'bid' | 'ask'
  pct: number
}) {
  return (
    <div className="grid grid-cols-[auto_1fr_auto] gap-x-2 text-xs font-mono py-0.5">
      <span className={side === 'bid' ? 'text-green-400' : 'text-red-400'}>{price}</span>
      <div className="relative">
        <div
          className={`absolute inset-y-0 left-0 ${side === 'bid' ? 'bg-green-400/20' : 'bg-red-400/20'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-gray-300 text-right">{quantity}</span>
    </div>
  )
}

function maxQty(entries: [string, string][]): number {
  return Math.max(...entries.map(([, qty]) => parseFloat(qty)))
}

export default function OrderBookTable({ book }: Props) {
  if (!book) return <p className="text-gray-500 text-xs">LOADING...</p>

  const maxAsk = maxQty(book.asks)
  const maxBid = maxQty(book.bids)

  return (
    <div>
      <div className="grid grid-cols-2 text-xs text-gray-500 mb-1">
        <span>Price</span>
        <span className="text-right">Quantity</span>
      </div>
      {book.asks.slice().reverse().map(([price, quantity]) => (
        <Row
          key={price}
          price={Number(price).toFixed(2)}
          quantity={Number(quantity).toFixed(5)}
          side="ask"
          pct={(parseFloat(quantity) / maxAsk) * 100}
        />
      ))}
      <div className="border-t border-gray-700 my-1" />
      {book.bids.map(([price, quantity]) => (
        <Row
          key={price}
          price={Number(price).toFixed(2)}
          quantity={Number(quantity).toFixed(5)}
          side="bid"
          pct={(parseFloat(quantity) / maxBid) * 100}
        />
      ))}
    </div>
  )
}