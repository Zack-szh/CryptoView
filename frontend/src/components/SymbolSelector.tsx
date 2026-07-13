interface Prop {
    symbols: string[]
    selected: string
    onChange: (s: string) => void
}

// given a list of symbols, return the selected symbol, then triggers onChange

export default function SymbolSelector({symbols, selected, onChange}: Prop) {
    return (
        <div className="flex gap-2 flex-wrap">
            {symbols.map((s) => (
        <button
            key={s}
            onClick={() => onChange(s)}
            className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
            s === selected
                ? 'bg-blue-600 text-white'
                : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
            }`}>
        
        {s}
        </button>))}
        </div>
    )
}