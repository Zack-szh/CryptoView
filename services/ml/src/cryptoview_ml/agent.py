import asyncio
import re
import sys

from langchain_core.messages import AIMessage, ToolMessage
from langchain.agents import create_agent

from .config import Settings, load_settings
from .llm import assert_supports_tool_calling, build_llm
from .tools import ALL_TOOLS

SYSTEM_PROMPT = "You are a crypto market analyst for the service CryptoView. " \
"Answer questions about the cryto market."

# extract content from think block, using regular expression pattern
THINK_BLOCK = re.compile(r"<think>.*?</think>", re.DOTALL)

def build_agent(settings: Settings):
    """ 
    Creates a ReAct loop and tools
    """
    llm = build_llm(settings)
    assert_supports_tool_calling(llm, settings.model)
    return create_agent(llm, ALL_TOOLS, system_prompt=SYSTEM_PROMPT)


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


async def ask(agent, question: str) -> list:
    """ 
    Note that we are using ainvoke, everything in tools should be async
    """
    result = await agent.ainvoke({"messages": [{"role": "user", "content": question}]})
    return result["messages"]


async def main():
    # here SOLANA is not in our testing database
    # agent should call get_symbols first to check whether data exists
    question = "What is BTC doing right now? give me a trend prediction on 5m interval"

    settings = load_settings() 
    agent = build_agent(settings)
    print(f"Question: {question}\n")
    messages = await ask(agent, question)

    print_trace(messages=messages)
    print(f"Content: {extract_text(messages[-1])}")

if __name__ == "__main__":
   asyncio.run(main())
    