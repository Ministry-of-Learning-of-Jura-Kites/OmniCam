from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env", env_file_encoding="utf-8", extra="ignore"  # Default to no file
    )

    dev_mode: bool = False

    # redis_host: str

    # redis_port: str

    # redis_req_topic: str

    # redis_res_topic: str

    nats_url: str

    req_topic_pattern: str

    res_topic_pattern: str

    req_topic_queue: str

    model_file_path: str


env_settings = Settings()
