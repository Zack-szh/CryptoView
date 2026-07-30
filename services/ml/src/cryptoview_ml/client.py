import os 
import time
from typing import Any

import httpx 
from dotenv import load_dotenv

load_dotenv() 

BASE_URL = os.getenv("API_BASE_URL", "http://localhost:8080/api/v1")

# /kline/:symbol now accept both "since" and "limit"
# /indicator/:symbol/series still takes "since" timstamp
# therefore we need to convert "since" into ms
INTERVAL_MS = {
    "1m": 60_000,
    "5m": 300_000,
    "15m": 900_000,
    "1h": 3_600_000,
    "4h": 14_400_000,
}


def since_ms(interval: str, bars: int) -> int:
    """Epoch-ms timestamp from `bars` intervals ago."""
    step = INTERVAL_MS.get(interval)
    if step is None:
        raise ValueError(f"unknown interval {interval!r}; use one of {list(INTERVAL_MS)}")
    return int(time.time() * 1000) - bars * step


async def _get(path: str, **params: Any) -> Any:
    """ GET on our Cryptpview API """
    clean = {k: v for k, v in params.items() if v is not None}

    async with httpx.AsyncClient(base_url=BASE_URL, timeout=10.0) as http: 
        resp = await http.get(path, params=clean)
        resp.raise_for_status()
        return resp.json()

# API handler
async def symbols() -> Any:
    return await _get("/symbols")

async def ticker(symbol: str, limit: int = 1) -> Any:
    return await _get(f"/ticker/{symbol}", limit=limit)

async def trades(symbol: str, limit: int = 10) -> Any:
    return await _get(f"/trade/{symbol}", limit=limit)

async def klines(symbol: str, interval: str = "5m", limit: int = 100) -> Any: 
    return await _get(f"/kline/{symbol}", interval=interval, limit=limit)

async def orderbook(symbol: str, limit: int = 20) -> Any:
    return await _get(f"/orderbook/{symbol}", limit=limit)

async def indicator_snapshot(symbol: str, interval: str = "5m") -> Any:
    return await _get(f"/indicator/{symbol}", interval=interval)

async def indicator_series(symbol: str, interval: str = "5m", bars: int = 200) -> Any:
     return await _get(
        f"/indicator/{symbol}/series",
        interval=interval,
        since=since_ms(interval, bars),
    )

# smoke test
if __name__ == "__main__":
    import asyncio

    def summarize(value: Any) -> str:
        if isinstance(value, list):
            return f"list[{len(value)}]"
        if isinstance(value, dict):
            return f"dict{sorted(value)}"
        return repr(value)

    async def smoke():
        syms = await symbols()
        print("symbols:   ", syms)
        if not syms:
            print("no symbols in the DB — is the market-data service running?")
            return

        s = syms[0]
        print("ticker:    ", summarize(await ticker(s)))
        print("trades:    ", summarize(await trades(s, 5)))
        print("klines:    ", summarize(await klines(s, "15m", 5)))
        print("orderbook: ", summarize(await orderbook(s, 5)))
        print("indicators:", summarize(await indicator_snapshot(s, "15m")))
        print("series:    ", summarize(await indicator_series(s, "15m", 10)))

    asyncio.run(smoke())