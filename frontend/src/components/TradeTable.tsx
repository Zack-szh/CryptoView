import type { Trade } from '../api/types'

interface Props {
  trades: Trade[]
}

export default function TradeTable({ trades }: Props) {
  return (
    <div className="overflow-y-auto max-h-64">
      <table className="w-full text-sm text-left">
        <thead className="text-gray-500 sticky top-0 bg-gray-900">
          <tr>
            <th className="pb-2">Time</th>
            <th className="pb-2">Price</th>
            <th className="pb-2">Qty</th>
            <th className="pb-2">Side</th>
          </tr>
        </thead>
        <tbody>
          {trades.map((t) => (
            <tr key={t.trade_id} className="border-t border-gray-800">
              <td className="py-1 text-gray-400">
                {new Date(t.time).toLocaleTimeString()}
              </td>
              <td className={`py-1 font-mono ${t.is_maker ? 'text-red-400' : 'text-green-400'}`}>
                {t.price.toFixed(2)}
              </td>
              <td className="py-1 text-gray-300">{t.quantity.toFixed(5)}</td>
              <td className="py-1 text-gray-500">{t.is_maker ? 'Sell' : 'Buy'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}