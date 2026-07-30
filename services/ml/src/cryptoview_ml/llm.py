from langchain_core.language_models import BaseChatModel
from langchain_core.tools import tool
# using chatmodels as opposed to legacy llm

from .config import Settings
from pydantic import SecretStr

def build_llm(settings: Settings) -> BaseChatModel:
    """
    Returns a BaseChatModel based on our config.
    Ollama, LM Studio, vLLM, mlx_lm, and llama.cpp's server all expose an OpenAI-compatible
    /v1 endpoint, so "local" is just ChatOpenAI pointed at a different base_url.
    Only Anthropic needs a client of its own.
    """

    if settings.provider == "anthropic":
        from langchain_anthropic import ChatAnthropic

        return ChatAnthropic(
            model_name=settings.model,
            api_key=SecretStr(settings.llm_api_key or ""),
            max_tokens=settings.max_tokens,
            timeout=None,
            stop=None,
        )

    from langchain_openai import ChatOpenAI

    return ChatOpenAI(
        model=settings.model,
        api_key=SecretStr(settings.llm_api_key or "not-needed"),  # local servers ignore it
        base_url=settings.llm_base_url,  # None → api.openai.com
        max_tokens=settings.max_tokens,
        timeout=None,
    )

@tool
def echo(value: str) -> str:
    """Echo back the given value, used for testing tool calling"""
    return value


def assert_supports_tool_calling(llm: BaseChatModel, model_name: str) -> None:
    """Verify the model emits structured tool calls, not prose that looks like one.

    This agent is entirely tool-driven: every number it reports has to come from a
    tool result. A model that cannot call tools does not fail loudly — it writes a
    plausible answer with invented prices. Catch it here, at startup, rather than
    shipping fabricated data to a user.

    """
    try:
        result = llm.bind_tools([echo]).invoke(
            "Call the echo tool with the value 'ping'. Do not reply with text."
        )
    except Exception as e:
        raise RuntimeError(f"Could not reach model {model_name!r}: {e}") from e

    if not getattr(result, "tool_calls", None):
        raise RuntimeError(
            f"Model {model_name!r} did not emit a tool call when explicitly asked to.\n"
            f"It replied with text instead: {str(result.content)[:200]!r}\n\n"
        )

if __name__ == "__main__":
    from .config import load_settings

    settings = load_settings()
    print(
        f"provider={settings.provider} model={settings.model} "
        f"base_url={settings.llm_base_url or '(provider default)'}"
    )

    llm = build_llm(settings)


    reply = llm.invoke("Introduce yourself")
    print("content:  ", repr(reply.content))
    print("finish:   ", reply.response_metadata.get("finish_reason"))
    print("extra:    ", reply.additional_kwargs)
    print("usage:    ", reply.usage_metadata)

    assert_supports_tool_calling(llm, settings.model)
    print("tools:  OK — model emitted a tool call")