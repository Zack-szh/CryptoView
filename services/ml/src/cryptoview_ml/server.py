# this file contains the HTTP server that exposes our chatbot to react frontend 

from contextlib import asynccontextmanager
from typing import Any

from fastapi import FastAPI, HTTPException
from langchain_core.messages import AIMessage, ToolMessage
from pydantic import BaseModel

from .agent import build_agent, extract_text
from .config import load_settings

# agent is built once on startup
_agent = None 

@asynccontextmanager
async def lifespan(app: FastAPI):
    global _agent
    settings = load_settings()
    print(f"building agent: provider={settings.provider} model={settings.model}")
    _agent = build_agent(settings)
    yield


app = FastAPI(lifespan=lifespan)

class ChatRequest(BaseModel):
    question: str 

class ToolCall(BaseModel):
    name: str
    args: dict[str, Any]
    result: str

class ChatResponse(BaseModel):
    answer: str
    tool_calls: list[ToolCall]


def collect_tool_calls(messages) -> list[ToolCall]:
    """ 
    Pairs each tool call with its result, linked by tool_call_id
    """
    results = {
        m.tool_call_id: str(m.content) for m in messages if isinstance(m, ToolMessage)
    }

    calls: list[ToolCall] = []
    for m in messages:
        if isinstance(m, AIMessage) and m.tool_calls:
            for c in m.tool_calls:
                if c["id"] == None:
                    raw = ""
                else:
                    raw = results.get(c["id"], "")
                calls.append(
                    ToolCall(
                        name=c["name"],
                        args=c["args"],
                        # kline payloads too big... 
                        result=raw[:800] + ("…" if len(raw) > 800 else ""),
                    )
                )
    return calls

@app.post("/agent/chat", response_model=ChatResponse)
async def chat(req: ChatRequest) -> ChatResponse:
    if _agent is None:
        raise HTTPException(status_code=503, detail="Agent not ready")

    try: 
        result = await _agent.ainvoke({"messages": [{"role": "user", "content": req.question}]})
    except Exception as e:
        raise HTTPException(status_code=502, detail="Failed to invoke agent {e}") from e

    messages = result["messages"]
    return ChatResponse(
        answer=extract_text(messages[-1]), 
        tool_calls=collect_tool_calls(messages)
    )

    