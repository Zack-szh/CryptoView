import asyncio
import re
import sys
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from uuid import uuid4

from langchain_core.messages import AIMessage, ToolMessage, HumanMessage
from langchain.agents import create_agent
from langgraph.checkpoint.base import BaseCheckpointSaver
from langgraph.checkpoint.postgres.aio import AsyncPostgresSaver

from .config import Settings, load_settings
from .llm import assert_supports_tool_calling, build_llm
from .tools import ALL_TOOLS

SYSTEM_PROMPT = "You are a crypto market analyst for the service CryptoView. " \
"Answer questions about the cryto market."

# extract content from think block, using regular expression pattern
THINK_BLOCK = re.compile(r"<think>.*?</think>", re.DOTALL)

def build_agent(settings: Settings, checkpointer: BaseCheckpointSaver):
    """ 
    Creates a ReAct loop and tools, along with AsyncPostgres checkpointer
    """
    llm = build_llm(settings)
    assert_supports_tool_calling(llm, settings.model)
    return create_agent(
        llm, 
        ALL_TOOLS,
        system_prompt=SYSTEM_PROMPT,
        checkpointer=checkpointer
    )


@asynccontextmanager
async def agent_session(settings: Settings) -> AsyncIterator: 
    """ 
    An agent backed by conversation history in postgres

    The saver owns a live connection to our postgres database, therefore we can't 
    simply return the agent and close this connection, that would mean the checkpointer 
    would point at nothing. Need to yield instead of return
    """

    async with AsyncPostgresSaver.from_conn_string(settings.database_url) as checkpointer:
        # creates checkpoint table if not exist, this is idempotent and runs every reboot
        # note that the saver owns the table, not in infra/migrations
        await checkpointer.setup() 
        yield build_agent(settings, checkpointer)



def extract_text(message) -> str:
    """
    Pulls text field out of AImessage
    Anthropic returns content as a list of typed block
    OpenAI compatiable endpoints return as plain text
    """
    content = message.content 

    if isinstance(content, str):
        text = content 
    else:
        text = "".join(
            block.get("text", "")
            for block in content
            if isinstance(block, dict) and block.get("type") == "text"
        )

    # gets rid of all content within thinkblock, only return content after
    return THINK_BLOCK.sub("", text).strip()


def print_trace(messages) -> None:
    """
    Prints every tool the agent called and tool's return value
    Traces the agent's tool call 
    """

    for m in messages:
        if isinstance(m, AIMessage) and m.tool_calls:   # tool requests
            for tool in m.tool_calls:
                print(f"   -> {tool['name']}({tool['args']})")
        elif isinstance(m, ToolMessage):                 # tool results
            preview = str(m.content).replace("\n", " ")[:160]
            print(f"  <- {m.name}: {preview}")


def latest_turn(messages: list) -> list: 
    """ 
    with checkpointer implemented, now messages includes all messages in a thread
    this function returns all messages starting from the last human message
    AKA: function returns all messages after the last user request
    """
    for i in range(len(messages)-1, -1, -1):
        if isinstance(messages[i], HumanMessage): 
            return messages[i:]
    return messages

async def ask(agent, question: str, thread_id: str | None = None) -> list:
    """ 
    Note that we are using ainvoke, everything in tools should be async

    thread_id is the key for checkpointer, same id means same conversation, new id 
    means new checkpoint, no id means stateless agent
    """

    config = {"configurable": {"thread_id": thread_id}} if thread_id else None
    result = await agent.ainvoke(
        {"messages": [{"role": "user", "content": question}]},
        config=config,
    )
    return result["messages"]


async def main():
    # the second question is the actual test: "what about ETH" is answerable 
    # if checkpoint works properly

    questions = [
        "What is BTC doing right now? You should analyze its momentum on the 5m chart.",
        "What about ETH? Same criteria as BTC.",
    ]

    settings = load_settings()
    thread_id = f"cli-{uuid4()}"

    async with agent_session(settings) as agent:
        print(f"thread: {thread_id}\n")
        for question in questions:
            print(f"Question: {question}")
            messages = await ask(agent, question, thread_id)
            print_trace(latest_turn(messages))
            print(f"Content: {extract_text(messages[-1])}\n")
    

if __name__ == "__main__":
   asyncio.run(main())
    