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

function cumulative(entries: [string, string][]): number[] {
  const result: number[] = []
  let sum = 0
  for (const [, qty] of entries) {
    sum += parseFloat(qty)
    result.push(sum)
  }
  return result
}

export default function OrderBookDepth({ book }: Props) {
  if (!book) return <p className="text-gray-500 text-xs">LOADING...</p>

  // asks are displayed reversed (lowest ask nearest spread) but cumulative
  // should accumulate away from spread, so compute on reversed array then reverse back
  const asksReversed = book.asks.slice().reverse()
  const askCumulative = cumulative(asksReversed).reverse()
  const bidCumulative = cumulative(book.bids)

  const maxAsk = askCumulative[0]  // worst ask has highest cumulative
  const maxBid = bidCumulative[bidCumulative.length - 1]  // worst bid has highest cumulative

  return (
    <div>
      <div className="grid grid-cols-[auto_1fr_auto] text-xs text-gray-500 mb-1 gap-x-2">
        <span>Price</span>
        <span />
        <span className="text-right">Quantity</span>
      </div>
      {asksReversed.map(([price, quantity], i) => (
        <Row
          key={price}
          price={Number(price).toFixed(2)}
          quantity={Number(quantity).toFixed(5)}
          side="ask"
          pct={(askCumulative[i] / maxAsk) * 100}
        />
      ))}
      <div className="border-t border-gray-700 my-1" />
      {book.bids.map(([price, quantity], i) => (
        <Row
          key={price}
          price={Number(price).toFixed(2)}
          quantity={Number(quantity).toFixed(5)}
          side="bid"
          pct={(bidCumulative[i] / maxBid) * 100}
        />
      ))}
    </div>
  )
}