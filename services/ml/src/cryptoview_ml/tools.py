from typing import Any, Literal 

import httpx
from langchain_core.tools import tool

from . import client

# only valid intervals 
Interval = Literal["1m", "5m", "15m", "1h", "4h"]

# another http clent wrapper for client.py
# this way if client.py fails, we still return a valid json instead of error
async def _call(coroutine) -> Any:
    """
    Runs a client call, if client call fails, still return data instead of exception. 
    A raised exception would crash the agent, we want the agent to know the call has failed, 
    instead of crashing or inventing data. 
    """
    try:
        return await coroutine 
    except httpx.HTTPStatusError as e: 
        return {"error": f"CryptoView API returned HTTP: {e.response.status_code}"}
    except httpx.RequestError as e:
        return {"error": f"Could not reach CryptoView API: {e}"}
    except ValueError as e:
        return {"error": str(e)}


# each tool corresponds to an API endpoint
@tool 
async def get_symbols() -> Any:
    """ List every trading symbol the database has data for. 
    Calls this before answering any question about a specific symbol to ensure 
    the data exists. If the symbol is not in this list, that means we do not have 
    data for this symbol.
    """
    return await _call(client.symbols())

@tool 
async def get_ticker(symbol: str, limit: int=1) -> Any: 
    """Get the most recent ticker rows for a symbol: last price, 24h open, high,
    low, volume, and trade count.

    Call this whenever the current price matters. `limit` is how many recent rows
    to return; 1 gives you just the latest.
    """
    return await _call(client.ticker(symbol, limit))

@tool
async def get_klines(symbol: str, interval: Interval = "15m", limit: int = 100) -> Any:
    """Get the last `limit` OHLCV candles for a symbol at the given interval.

    Use this to read price action and trend over time. Prefer 15m or 5m for
    multi-hour context; 1m is noisy and rarely worth the tokens. Keep `limit`
    under ~200 unless you genuinely need the history. Klines should give you the 
    most up-to-date and accurate price movement.
    """
    return await _call(client.klines(symbol, interval, limit))

@tool 
async def get_trades(symbol: str, limit: int=10) -> Any:
    """Get the most recent individual trades for a symbol.

    Useful only for very short-term buying vs selling pressure. For anything
    longer than a few minutes, use get_klines instead. Also note that the API 
    we are getting data from does not have frequent trade activities.
    """
    return await _call(client.trades(symbol, limit))
    
@tool
async def get_orderbook(symbol: str, limit: int = 20) -> Any:
    """Get the current live order book (bids and asks) for a symbol.

    This is a real-time call straight to the exchange, so it reflects the book
    right now rather than stored history. Use it to judge immediate liquidity
    and the spread.
    """
    return await _call(client.orderbook(symbol, limit))


@tool
async def get_indicator_snapshot(symbol: str, interval: Interval = "15m") -> Any:
    """Get the latest value of every technical indicator for a symbol: SMA, EMA,
    RSI, MACD, Bollinger Bands, VWAP, and realized volatility.

    This is cheap and is usually all you need. Reach for it before get_klines
    when the question is about momentum or trend rather than raw price.
    """
    return await _call(client.indicator_snapshot(symbol, interval))


ALL_TOOLS = [
    get_symbols,
    get_ticker,
    get_trades,
    get_klines,
    get_orderbook,
    get_indicator_snapshot,
]