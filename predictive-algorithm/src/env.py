from pydantic_settings import BaseSettings, SettingsConfigDict

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env", env_file_encoding="utf-8", extra="ignore"
    )

    dev_mode: bool = False
    nats_url: str
    req_topic_pattern: str
    req_topic_queue: str
    model_file_path: str


env_settings = Settings()
