import os 
from dataclasses import dataclass 

from dotenv import load_dotenv

load_dotenv() 

VALID_PROVIDERS=("anthropic", "openai", "local")

@dataclass(frozen=True)
class Settings:
    provider: str
    model: str
    llm_base_url: str | None
    llm_api_key: str | None
    max_tokens: int
    # 8/7/2026L Database_url added, for persisitng langgraph checkpoints
    database_url: str


def load_settings() -> Settings:
    """ 
    Read LLM settings from local env
    if any config is missing, raise runtime error 
    """

    provider = os.getenv("LLM_PROVIDER")
    if not provider:
        raise RuntimeError(
            f"LLM_PROVIDER is not set, Choose one of {', '.join(VALID_PROVIDERS)}"
        )
    
    if provider not in VALID_PROVIDERS:
        raise RuntimeError(
            f"Invalid LLM_PROVIDER: {provider}"
        )

    model = os.getenv("LLM_MODEL")
    if not model:
        raise RuntimeError("LLM_MODEL is not set (e.g. claude-opus-5, or qwen3:14b)")

    llm_base_url = os.getenv("LLM_BASE_URL")
    if provider == "local" and not llm_base_url:
        raise RuntimeError(
            "LLM_PROVIDER=local requires LLM_BASE_URL "
            "(e.g. http://localhost:11434/v1 for Ollama)"
        )

    llm_api_key = os.getenv("LLM_API_KEY")
    if provider in ("anthropic", "openai") and not llm_api_key:
        raise RuntimeError(f"LLM_PROVIDER={provider} requires LLM_API_KEY")

    # never inherit the server's own default: mlx_lm.server caps at 512, which a
    # thinking model spends entirely inside <think> before writing a single word
    raw_max_tokens = os.getenv("LLM_MAX_TOKENS", "8192")
    try:
        max_tokens = int(raw_max_tokens)
    except ValueError:
        raise RuntimeError(f"LLM_MAX_TOKENS={raw_max_tokens} is not an integer") from None

    database_url = os.getenv("DATABASE_URL", "")
    if database_url == "": 
        raise RuntimeError("invalid env DATABASE_URL")

    return Settings(
        provider=provider,
        model=model,
        llm_base_url=llm_base_url,
        llm_api_key=llm_api_key,
        max_tokens=max_tokens,
        database_url=database_url,
    )