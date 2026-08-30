"""Centralized configuration loaded from environment variables.

Using pydantic-settings gives us validation at startup: a missing or
malformed required variable fails fast with a clear error instead of an
obscure runtime exception later - this is deliberate, so "incorrect
environment variable" is a reproducible, easy-to-diagnose failure mode.
"""
from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    port: int = 8000
    log_level: str = "info"
    environment: str = "development"

    db_host: str = "localhost"
    db_port: int = 5432
    db_name: str = "shopstream_products"
    db_user: str = "shopstream"
    db_password: str = ""
    db_pool_size: int = 5
    db_pool_max_overflow: int = 10
    db_connect_timeout_seconds: int = 5

    default_page_size: int = 20
    max_page_size: int = 100

    @property
    def database_url(self) -> str:
        return (
            f"postgresql+psycopg2://{self.db_user}:{self.db_password}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}"
        )


@lru_cache
def get_settings() -> Settings:
    return Settings()
