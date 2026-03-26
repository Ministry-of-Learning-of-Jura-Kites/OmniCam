from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

    dev_mode: bool = False

    redis_host: str

    redis_port: str

    redis_req_topic: str

    redis_res_topic: str

    model_file_path: str


env_settings = Settings()
