import type { OrderBook } from "../api/types";

interface Props {
    book: OrderBook | null 
}


// this represents one row in the order book

function Row({price, quantity, side}: {price: string; quantity: string; side: 'bid' | 'ask'}) { 
    return (
        <div className="grid grid-cols-2 text-xs font-mono py-0.5">
            <span className={side == 'bid' ? 'text-green-400' : 'text-red-400'}>{price}</span> 
            <span className="text-gray-300 text-right">{quantity}</span>
        </div>
    )
}

// for asks, we do book.asks.slice().reveres() because the lowest ask is closest to spread
// AKA: lowest ask should always be right above highest bid
export default function OrderBook({book} : Props) {
    if (!book) return <p className="text-gray-500 text-xs">LOADING...</p>

    return (
        <div>
            <div className="grid grid-cols-2 text-xs text-gray-500 mb-1">
                <span>Price</span>
                <span className="text-right">Quantity</span>
            </div>
            {book.asks.slice().reverse().map(([price, quantity]) => (
            <Row key={price} price={price} quantity={quantity} side="ask" />
        ))}
        <div className="border-t border-gray-700 my-1" />
        {book.bids.map(([price, quantity]) => (
            <Row key={price} price={price} quantity={quantity} side="bid" />))}
    </div>
  )
}