# CryptoView

Real-time cryptocurrency market data platform. Streams live trades, tickers, and klines from Binance.us into PostgreSQL, serves them over a REST API, computes technical indicators on the fly, and surfaces everything through a React dashboard with an LLM-powered analyst agent on top.


![Demo](tour.gif)


## Features

- **Live market data ingestion** — WebSocket streams from Binance (`@trade`, `@ticker`, `@kline_*`, `@depth20`) persisted to PostgreSQL with automatic reconnect and gap backfill via the REST API
- **REST API** — Go + Gin service exposing symbols, tickers, trades, klines, order book, and computed indicators
- **Technical indicators** — SMA, EMA, RSI, MACD, Bollinger Bands, VWAP, and realized volatility, available as a snapshot or a full historical series
- **Interactive dashboard** — React + TypeScript + Vite frontend with a live candlestick chart, order book, trade tape, and ticker panel
- **LLM analyst agent** — LangChain/LangGraph agent (Anthropic, OpenAI, or any local OpenAI-compatible endpoint) that answers questions about the market using only live tool-sourced data, with conversation history persisted per session in Postgres. Every single tool call is logged and retraceable within the agent interface. 
![Agent Tool Calling](p4.png)

## Architecture

```
Binance WebSocket
       │
       ▼
market-data service (Go)  ──stream + backfill──▶  PostgreSQL
                                                        ▲
                                                        │
                                    api service (Go) ───┘
                                    :8080  /api/v1/*
                                        ▲          ▲
                                        │          │
                          Vite dev server         ml agent service (FastAPI)
                          :3000  (browser)        :8000  /agent/chat
                                        │                │
                                        │                ├──▶ Go API (tool calls)
                                        │                └──▶ LLM provider (Anthropic / OpenAI / local)
                                        ▼
                                     React UI
```

## Project layout

```
CryptoView/
├── services/
│   ├── market-data/     # Binance WebSocket ingestion, DB writes, backfill/maintainer
│   ├── api/              # REST API (Gin) over PostgreSQL
│   ├── indicators/       # Technical indicator calculations
│   └── ml/                # LangChain/LangGraph agent + FastAPI chat server (Python)
├── frontend/              # React + TypeScript + Vite dashboard
├── infra/migrations/      # SQL schema migrations
└── docker-compose.yml     # PostgreSQL service
```

## Tech stack

| Layer | Technology |
|---|---|
| Market data ingestion | Go, `gorilla/websocket`, `pgx` |
| API | Go, `gin-gonic/gin` |
| Database | PostgreSQL 18 |
| Indicators | Go |
| Agent | Python, LangChain, LangGraph, FastAPI, `httpx` |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, `lightweight-charts` |

## Prerequisites

- [Go](https://go.dev/) 1.26+
- [Node.js](https://nodejs.org/) + npm
- [Python](https://www.python.org/) 3.12+ with [`uv`](https://docs.astral.sh/uv/)
- [Docker](https://www.docker.com/) (for PostgreSQL)
- An LLM backend: an Anthropic/OpenAI API key, or a local OpenAI-compatible server (e.g. `mlx_lm.server`, Ollama, LM Studio)

## Configuration

Create a `.env` file in the project root:

```bash
# Database
DATABASE_URL=postgres://cryptoview:cryptoview@localhost:5432/cryptoview

# Market data ingestion
SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
HISTORY_DAYS=90

# API
API_PORT=8080
API_BASE_URL=http://localhost:8080/api/v1

# LLM agent
LLM_PROVIDER=anthropic        # anthropic | openai | local
LLM_MODEL=claude-opus-5
LLM_BASE_URL=                 # required for LLM_PROVIDER=local, e.g. http://127.0.0.1:8081/v1
LLM_API_KEY=
```

## Getting started

**1. Start PostgreSQL**

```bash
docker-compose up
```

**2. Run the market data ingestion service** (streams live data into the DB)

```bash
go run ./services/market-data/cmd/stream/main.go
```

**3. Run the API service**

```bash
go run ./services/api/cmd/api/main.go
```

**4. Run the agent service**

```bash
cd services/ml
uv run uvicorn cryptoview_ml.server:app --port 8000 --reload
```

**5. Run the frontend**

```bash
cd frontend
npm install
npm run dev
```

Open the dashboard at [http://localhost:3000](http://localhost:3000).

## API reference

All endpoints are served under `/api/v1`.

| Method | Path | Description |
|---|---|---|
| GET | `/symbols` | List all tracked symbols |
| GET | `/ticker/:symbol` | Latest ticker snapshots (`limit`, default 10, max 500) |
| GET | `/trade/:symbol` | Recent trades (`limit`, default 10, max 500) |
| GET | `/kline/:symbol` | Candlestick data (`interval`, `limit`/`since`) |
| GET | `/orderbook/:symbol` | Live order book depth |
| GET | `/indicator/:symbol` | Latest indicator snapshot |
| GET | `/indicator/:symbol/series` | Historical indicator series |

The agent service exposes:

| Method | Path | Description |
|---|---|---|
| POST | `/agent/chat` | Ask the analyst agent a question (`question`, `session_id`) |

## Useful commands

```bash
# Reset the database (drops all data)
docker-compose down -v

# Open a psql shell
docker exec -it cryptoview-postgres-1 psql -U cryptoview -d cryptoview

# Serve a local model for the agent (example: mlx-lm)
mlx_lm.server --port 8081 --model <model-path>
```



## License

MIT License

Copyright (c) 2026 Zack Sun
