import { useEffect, useRef, useState } from 'react'
import { askAgent } from '../api/client'
import type { AgentToolCall } from '../api/types'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface Turn {
  question: string
  answer?: string
  toolCalls?: AgentToolCall[]
  error?: string
}

const SUGGESTIONS = [
  'What is BTCUSDT doing right now?',
  'Compare BTCUSDT and ETHUSDT momentum',
  'Is ETHUSDT overbought?',
]

// tool calls are shown collapsed
function ToolCalls({ calls }: { calls: AgentToolCall[] }) {
  if (!calls.length) return null

  return (
    <div className="mt-3 space-y-1">
      {calls.map((c, i) => (
        <details key={i} className="text-xs">
          <summary className="cursor-pointer text-gray-500 hover:text-gray-300">
            {c.name}({Object.entries(c.args).map(([k, v]) => `${k}=${v}`).join(', ')})
          </summary>
          <pre className="mt-1 max-h-40 overflow-auto rounded bg-gray-950 p-2 text-gray-400">
            {c.result}
          </pre>
        </details>
      ))}
    </div>
  )
}

// the model can take 30s+ locally, a static spinner reads as broken
function Elapsed() {
  const [secs, setSecs] = useState(0)
  useEffect(() => {
    const t = setInterval(() => setSecs((s) => s + 1), 1000)
    return () => clearInterval(t)
  }, [])
  return <span className="text-sm text-gray-500">thinking… {secs}s</span>
}

export default function AgentChat() {
  const [input, setInput] = useState('')
  const [turns, setTurns] = useState<Turn[]>([])
  const [busy, setBusy] = useState(false)
  // the agent's checkpointer loads history by sessionId
  const [sessionId, setSessionId] = useState(() => crypto.randomUUID())
  const bottom = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottom.current?.scrollIntoView({ behavior: 'smooth' })
  }, [turns, busy])

  // a new id IS the reset — the old thread stays in postgres, orphaned and unreachable
  function newChat() {
    if (busy) return
    setSessionId(crypto.randomUUID())
    setTurns([])
    setInput('')
  }


  async function send(question: string) {
    if (!question || busy) return

    setInput('')
    setBusy(true)
    setTurns((t) => [...t, { question }])

    // patch the last turn in place, whichever way it resolves
    const patch = (fields: Partial<Turn>) =>
      setTurns((t) => t.map((turn, i) => (i === t.length - 1 ? { ...turn, ...fields } : turn)))

    try {
      const reply = await askAgent(question, sessionId)
      patch({ answer: reply.answer, toolCalls: reply.tool_calls })
    } catch (err) {
      patch({ error: err instanceof Error ? err.message : String(err) })
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex h-[32rem] flex-col">
      <div className="mb-2 flex items-center justify-between">
        <span className="font-mono text-xs text-gray-600">session {sessionId}</span>
        <button
          onClick={newChat}
          disabled={busy || turns.length === 0}
          className="rounded px-2 py-1 text-xs text-white-500 hover:bg-white-800 hover:text-white-300 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          New chat
        </button>
      </div>

      <div className="flex-1 space-y-4 overflow-y-auto pr-1">
        {turns.length === 0 && (
          <div className="space-y-2">
            <p className="text-sm text-gray-500">Ask about any symbol in the database.</p>
            {SUGGESTIONS.map((s) => (
              <button
                key={s}
                onClick={() => send(s)}
                className="block w-full rounded-lg bg-gray-800 px-3 py-2 text-left text-sm text-gray-300 hover:bg-gray-700"
              >
                {s}
              </button>
            ))}
          </div>
        )}

        {turns.map((turn, i) => (
          <div key={i} className="space-y-2">
            <div className="ml-auto w-fit max-w-[85%] rounded-lg bg-blue-600 px-3 py-2 text-sm text-white">
              {turn.question}
            </div>

            {turn.error && (
              <div className="text-sm text-red-400">{turn.error}</div>
            )}

            {turn.answer && (
              <div className="rounded-lg bg-gray-800 px-3 py-2">
                <div className="prose prose-invert prose-sm max-w-none
                                prose-p:my-2 prose-ul:my-2 prose-li:my-0.5
                                prose-headings:mt-3 prose-headings:mb-1
                                prose-strong:text-white
                                prose-code:text-blue-300 prose-code:before:content-none prose-code:after:content-none">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{turn.answer}</ReactMarkdown>
                </div>
                <ToolCalls calls={turn.toolCalls ?? []} />
              </div>
            )}
          </div>
        ))}

        {busy && <Elapsed />}
        <div ref={bottom} />
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          send(input.trim())
        }}
        className="mt-3 flex gap-2"
      >
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask about the market…"
          disabled={busy}
          className="flex-1 rounded-lg bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={busy || !input.trim()}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-40"
        >
          Ask
        </button>
      </form>
    </div>
  )
}